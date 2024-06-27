package db

import "api/internal/models"

type Database interface {
	GetAllPosts() ([]*models.Post, error)
	GetPostByID(id string) (*models.Post, error)
	CreatePost(title string, content string, commentsEnabled bool) (*models.Post, error)
	CreateComment(postId string, parentId *string, content string) (*models.Comment, error)
	SubscribeToComments(postId string) (<-chan *models.Comment, error)
	NotifySubscribers(postId string, comment *models.Comment)
}
