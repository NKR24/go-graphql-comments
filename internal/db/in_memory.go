package db

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"api/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type InMemoryDB struct {
	posts         map[string]*models.Post
	comments      map[string]*models.Comment
	subscriptions map[string][]chan *models.Comment
	Redis         *redis.Client
	sync.RWMutex
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		posts:         make(map[string]*models.Post),
		comments:      make(map[string]*models.Comment),
		subscriptions: make(map[string][]chan *models.Comment),
	}
}

func (db *InMemoryDB) GetAllPosts() ([]*models.Post, error) {
	db.RLock()
	defer db.RUnlock()
	var posts []*models.Post
	for _, post := range db.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (db *InMemoryDB) GetPostByID(id string) (*models.Post, error) {
	db.RLock()
	defer db.RUnlock()
	post, exists := db.posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}
	return post, nil
}

func (db *InMemoryDB) CreatePost(title string, content string, commentsEnabled bool) (*models.Post, error) {
	db.Lock()
	defer db.Unlock()
	id := uuid.New().String()
	post := &models.Post{
		ID:              id,
		Title:           title,
		Content:         content,
		CommentsEnabled: commentsEnabled,
		Comments:        []models.Comment{},
	}
	db.posts[id] = post
	return post, nil
}

func (db *InMemoryDB) CreateComment(postId string, parentId *string, content string) (*models.Comment, error) {
	db.Lock()
	defer db.Unlock()
	post, exists := db.posts[postId]
	if !exists || !post.CommentsEnabled {
		return nil, fmt.Errorf("cannot add comment")
	}
	id := uuid.New().String()
	comment := &models.Comment{
		ID:       id,
		PostID:   postId,
		ParentID: parentId,
		Content:  content,
		Children: []models.Comment{},
	}
	db.comments[id] = comment
	if parentId != nil {
		parentComment, exists := db.comments[*parentId]
		if exists {
			parentComment.Children = append(parentComment.Children, *comment)
		}
	} else {
		post.Comments = append(post.Comments, *comment)
	}
	db.NotifySubscribers(postId, comment)
	return comment, nil
}

func (db *InMemoryDB) SubscribeToComments(postId string) (<-chan *models.Comment, error) {
	ch := make(chan *models.Comment)
	go func() {
		pubsub := db.Redis.Subscribe(context.Background(), fmt.Sprintf("post:%s:comments", postId))
		defer pubsub.Close()

		for msg := range pubsub.Channel() {
			var comment models.Comment
			if err := json.Unmarshal([]byte(msg.Payload), &comment); err == nil {
				ch <- &comment
			}
		}
	}()
	return ch, nil
}

func (db *InMemoryDB) NotifySubscribers(postId string, comment *models.Comment) {
	data, err := json.Marshal(comment)
	if err != nil {
		return
	}
	db.Redis.Publish(context.Background(), fmt.Sprintf("post:%s:comments", postId), data)
}
