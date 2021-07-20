package main

// Post ...
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

const (
	userID       = "7"
	POSTS_URL    = "https://jsonplaceholder.typicode.com/posts?userId="
	COMMENTS_URL = "https://jsonplaceholder.typicode.com/comments?postId="
)

func main() {
}
