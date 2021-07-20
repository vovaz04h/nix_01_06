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

	ERR_DB_CONNECT     = 1
	ERR_HTTP_GET       = 2
	ERR_HTTP_BODY      = 3
	ERR_JSON_UNMARSHAL = 4
)

var (
	db *sql.DB
)

func writeCommentToDB(comment Comment, wgComment sync.WaitGroup) {
	// write comments to database
	err := db.Ping()
	if err != nil {
		log.Println(err)
		return
	}

	// INSERT INTO comments (in, postid, name, email, body) VALUES ()

	wgComment.Done()
	return
}

func writePostToDB(post Post, wgPosts sync.WaitGroup) {
	// write posts to database
	err := db.Ping()
	if err != nil {
		log.Println(err)
		return
	}

	// INSERT INTO posts (id, userid, title, body) VALUES ()

	// get comments with postId
	resp, err := http.Get(COMMENTS_URL + strconv.Itoa(post.ID))
	if err != nil {
		log.Println(err)
		return
	}

	// close http.response.body
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var comments []Comment
	err = json.Unmarshal(body, &comments)
	if err != nil {
		log.Println(err)
		return
	}

	var wgComments sync.WaitGroup

	// write comments to database
	for _, comment := range comments {
		wgComments.Add(1)
		go writeCommentToDB(comment, wgComments)
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

	// close database connect
	defer db.Close()

	// get posts with userID=7
	resp, err := http.Get(POSTS_URL + userID)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_HTTP_GET)
	}

	// close http.response.body
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_HTTP_BODY)
	}

	var posts []Post
	err = json.Unmarshal(body, &posts)
	if err != nil {
		log.Println(err)
		os.Exit(ERR_JSON_UNMARSHAL)
	}

	var wgPosts sync.WaitGroup

	// write posts to database
	for _, post := range posts {
		wgPosts.Add(1)
		go writePostToDB(post, wgPosts)
	}
	wgPosts.Wait()
}
