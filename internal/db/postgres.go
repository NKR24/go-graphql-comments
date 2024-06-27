package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"api/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	conn  *sql.DB
	Redis *redis.Client
}

func NewPostgresDB(connStr string, redisAddr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &PostgresDB{conn: db, Redis: redisClient}, nil
}

func (db *PostgresDB) GetAllPosts() ([]*models.Post, error) {
	rows, err := db.conn.Query("SELECT id, title, content, comments_enabled FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CommentsEnabled); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (db *PostgresDB) GetPostByID(id string) (*models.Post, error) {
	var post models.Post
	err := db.conn.QueryRow("SELECT id, title, content, comments_enabled FROM posts WHERE id = $1", id).
		Scan(&post.ID, &post.Title, &post.Content, &post.CommentsEnabled)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (db *PostgresDB) CreatePost(title string, content string, commentsEnabled bool) (*models.Post, error) {
	id := uuid.New().String()
	_, err := db.conn.Exec("INSERT INTO posts (id, title, content, comments_enabled) VALUES ($1, $2, $3, $4)",
		id, title, content, commentsEnabled)
	if err != nil {
		return nil, err
	}
	return &models.Post{
		ID:              id,
		Title:           title,
		Content:         content,
		CommentsEnabled: commentsEnabled,
		Comments:        []models.Comment{},
	}, nil
}

func (db *PostgresDB) CreateComment(postId string, parentId *string, content string) (*models.Comment, error) {
	id := uuid.New().String()
	_, err := db.conn.Exec("INSERT INTO comments (id, post_id, parent_id, content) VALUES ($1, $2, $3, $4)",
		id, postId, parentId, content)
	if err != nil {
		return nil, err
	}
	comment := &models.Comment{
		ID:       id,
		PostID:   postId,
		ParentID: parentId,
		Content:  content,
		Children: []models.Comment{},
	}
	db.NotifySubscribers(postId, comment)
	return comment, nil
}

func (db *PostgresDB) SubscribeToComments(postId string) (<-chan *models.Comment, error) {
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

func (db *PostgresDB) NotifySubscribers(postId string, comment *models.Comment) {
	data, err := json.Marshal(comment)
	if err != nil {
		return
	}
	db.Redis.Publish(context.Background(), fmt.Sprintf("post:%s:comments", postId), data)
}
