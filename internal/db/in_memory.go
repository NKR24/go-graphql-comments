package db

import (
	"api/internal/models"
	"sync"
)

type InMemoryDB struct {
	posts    map[string]models.Post
	comments map[string]models.Comment
	sync.RWMutex
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		posts:    make(map[string]models.Post),
		comments: make(map[string]models.Comment),
	}
}
