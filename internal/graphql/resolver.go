package graphql

import (
	"api/internal/db"
	"api/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type Resolver struct {
	DB    db.Database
	Redis *redis.Client
}

func (r *Resolver) Posts(ctx context.Context) ([]*models.Post, error) {
	return r.DB.GetAllPosts()
}

func (r *Resolver) Post(ctx context.Context, id string) (*models.Post, error) {
	return r.DB.GetPostByID(id)
}

func (r *Resolver) CreatePost(ctx context.Context, title string, content string, commentsEnabled bool) (*models.Post, error) {
	return r.DB.CreatePost(title, content, commentsEnabled)
}

func (r *Resolver) CreateComment(ctx context.Context, postId string, parentId *string, content string) (*models.Comment, error) {
	comment, err := r.DB.CreateComment(postId, parentId, content)
	if err != nil {
		return nil, err
	}
	r.NotifySubscribers(postId, comment)
	return comment, nil
}

func (r *Resolver) CommentAdded(ctx context.Context, postId string) (<-chan *models.Comment, error) {
	ch := make(chan *models.Comment)
	go func() {
		pubsub := r.Redis.Subscribe(ctx, fmt.Sprintf("post:%s:comments", postId))
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

func (r *Resolver) NotifySubscribers(postId string, comment *models.Comment) {
	data, err := json.Marshal(comment)
	if err != nil {
		return
	}
	r.Redis.Publish(context.Background(), fmt.Sprintf("post:%s:comments", postId), data)
}
