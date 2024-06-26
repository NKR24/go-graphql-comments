package models

type Post struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	CommentsEnabled bool      `json:"commentsEnabled"`
	Comments        []Comment `json:"comments"`
}
