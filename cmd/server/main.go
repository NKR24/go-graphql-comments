package main

import (
	"log"
	"os"

	"api/internal/db"
	"api/internal/server"
	"github.com/go-redis/redis/v8"
)

func main() {
	var database db.Database
	storageType := os.Getenv("STORAGE_TYPE")
	redisAddr := os.Getenv("REDIS_ADDR")

	if storageType == "postgres" {
		connStr := os.Getenv("DATABASE_URL")
		pgDB, err := db.NewPostgresDB(connStr, redisAddr)
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		database = pgDB
	} else {
		inMemDB := db.NewInMemoryDB()
		redisClient := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		inMemDB.Redis = redisClient
		database = inMemDB
	}

	server.StartServer(database)
}
