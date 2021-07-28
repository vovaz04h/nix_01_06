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
	POSTS_URL    = "https://jsonplaceholder.typicode.com/posts?userId="
	COMMENTS_URL = "https://jsonplaceholder.typicode.com/comments?postId="

	ERR_DB_CONNECT = iota + 1
	ERR_DB_STMT
	ERR_GET_POSTS
)

var (
	db *sql.DB
)

func writeCommentToDB(comment Comment, stmtCommentSave *sql.Stmt) error {
	// write comments to database
	err := db.Ping()
	if err != nil {
		return err
	}

	// INSERT INTO comments (id, postid, name, email, body) VALUES ()
	_, err = stmtCommentSave.Exec(comment.ID, comment.PostID, comment.Name, comment.Email, comment.Body)
	if err != nil {
		return err
	}

	return nil
}

func writePostToDB(post Post, stmtPostSave *sql.Stmt, stmtCommentSave *sql.Stmt) error {

	// write posts to database
	err := db.Ping()
	if err != nil {
		return err
	}

	// INSERT INTO posts (id, userid, title, body) VALUES ()
	_, err = stmtPostSave.Exec(post.ID, post.UserID, post.Title, post.Body)
	if err != nil {
		return err
	}

	// get comments with postId
	comments, err := getComments(strconv.Itoa(post.ID))
	if err != nil {
		return err
	}

	// write comments to database
	wgComments := new(sync.WaitGroup)
	for _, comment := range comments {
		wgComments.Add(1)
		go func(comment Comment) {
			err := writeCommentToDB(comment, stmtCommentSave)
			if err != nil {
				log.Println(err)
			}
			wgComments.Done()
		}(comment)
	}

	wgComments.Wait()
	return nil
}

func getComments(postID string) (comments []Comment, err error) {
	// get comments with postId
	resp, err := http.Get(COMMENTS_URL + postID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// json to struct
	err = json.Unmarshal(body, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func getPosts(userID string) (posts []Post, err error) {
	// get posts with userID
	resp, err := http.Get(POSTS_URL + userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// json to struct
	err = json.Unmarshal(body, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func main() {
	// connect to database
	db, err := sql.Open("mysql", "[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]")
	if err != nil {
		log.Println(err)
		os.Exit(ERR_DB_CONNECT)
	}
	defer db.Close()

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

	// get Posts with userID=7
	userID := "7"
	posts, err := getPosts(userID)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_GET_POSTS)
	}

	// write posts to database
	wgPosts := new(sync.WaitGroup)
	for _, post := range posts {
		wgPosts.Add(1)
		go func(post Post) {
			err := writePostToDB(post, stmtPostSave, stmtCommentSave)
			if err != nil {
				log.Println(err)
			}
			wgPosts.Done()
		}(post)
	}

	wgPosts.Wait()
}
