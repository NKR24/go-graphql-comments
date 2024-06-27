package server

import (
	"log"
	"net/http"
	"os"
	"time"

	"api/internal/db"
	"api/internal/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

func NewServer(database db.Database, redisClient *redis.Client) *http.Server {
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{DB: database, Redis: redisClient}}))

	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
}

func StartServer(database db.Database) {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	srv := NewServer(database, redisClient)
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv.Handler)

	log.Printf("connect to http://localhost:8080/ for GraphQL playground")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
