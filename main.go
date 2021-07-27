package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// Post type
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Comment type
type Comment struct {
	PostID int    `json:"postId"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

const (
	userID       = "7"
	POSTS_URL    = "https://jsonplaceholder.typicode.com/posts?userId="
	COMMENTS_URL = "https://jsonplaceholder.typicode.com/comments?postId="

	ERR_DB_CONNECT = iota + 1
	ERR_DB_STMT
	ERR_HTTP_GET
	ERR_HTTP_BODY
	ERR_JSON_UNMARSHAL
)

var (
	db *sql.DB
)

func writeCommentToDB(comment Comment, stmtCommentSave *sql.Stmt, wgComment sync.WaitGroup) {
	// write comments to database
	err := db.Ping()
	if err != nil {
		log.Println(err)
		return
	}

	// INSERT INTO comments (id, postid, name, email, body) VALUES ()
	_, err = stmtCommentSave.Exec(comment.ID, comment.PostID, comment.Name, comment.Email, comment.Body)
	if err != nil {
		log.Panicln(err)
		return
	}

	wgComment.Done()
	return
}

func writePostToDB(post Post, stmtPostSave *sql.Stmt, stmtCommentSave *sql.Stmt, wgPosts sync.WaitGroup) {
	// write posts to database
	err := db.Ping()
	if err != nil {
		log.Println(err)
		return
	}

	// INSERT INTO posts (id, userid, title, body) VALUES ()
	_, err = stmtPostSave.Exec(post.ID, post.UserID, post.Title, post.Body)
	if err != nil {
		log.Panicln(err)
		return
	}

	// get comments with postId
	resp, err := http.Get(COMMENTS_URL + strconv.Itoa(post.ID))
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// json to struct
	var comments []Comment
	err = json.Unmarshal(body, &comments)
	if err != nil {
		log.Println(err)
		return
	}

	// write comments to database
	var wgComments sync.WaitGroup
	for _, comment := range comments {
		wgComments.Add(1)
		go writeCommentToDB(comment, stmtCommentSave, wgComments)
	}

	wgComments.Wait()
	wgPosts.Done()
	return
}

func main() {
	// connect to database
	db, err := sql.Open("mysql", "[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")
	if err != nil {
		log.Println(err)
		//		os.Exit(ERR_DB_CONNECT)
	}
	defer db.Close()

	// get posts with userID=7
	resp, err := http.Get(POSTS_URL + userID)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_HTTP_GET)
	}
	defer resp.Body.Close()

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_HTTP_BODY)
	}

	// json to struct
	var posts []Post
	err = json.Unmarshal(body, &posts)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_JSON_UNMARSHAL)
	}

	// sqlPreparePostSave
	qs := "INSERT INTO posts (id, userid, title, body) VALUES (?, ?, ?, ?)"
	stmtPostSave, err := db.Prepare(qs)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_DB_STMT)
	}
	defer stmtPostSave.Close()

	// sqlPrerareCommentSave
	qs = "INSERT INTO comments (id, postid, name, email, body) VALUES (?, ?, ?, ?, ?)"
	stmtCommentSave, err := db.Prepare(qs)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_DB_STMT)
	}
	defer stmtCommentSave.Close()

	// write posts to database
	var wgPosts sync.WaitGroup
	for _, post := range posts {
		wgPosts.Add(1)
		go writePostToDB(post, stmtPostSave, stmtCommentSave, wgPosts)
	}

	wgPosts.Wait()
}
