package models

type Comment struct {
	ID       string    `json:"id"`
	PostID   string    `json:"postId"`
	ParentID *string   `json:"parentId"`
	Content  string    `json:"content"`
	Children []Comment `json:"children"`
}
