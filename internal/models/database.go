package models

type DataBase interface {
	GetAllPosts() ([]*Post, error)
	GetPostById(id string) ([]*Post, error)
	CreatePost(title string, content string, commentsEnabled string) ([]*Post, error)
	CreareComment(postId string, parentId *string, content string) ([]*Comment, error)
	SubscribeToComments(postId string) (<-chan *Comment, error)
	NotifySubscribers(postId string, comment *Comment)
}
