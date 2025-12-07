package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	// 1. Config Fail fast if critical variables are missing.
	// In a bank, we don't start if we don't know where the db is
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisAddr == "" {
		log.Fatal("Error: REDIS_ADDR environment variable is required")
	}

	// 2. Initializing Client Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // Empty string if there is no pass (insecure, but allowed in dev local initial)
		DB:       0,
	})

	// 3. Infrastructure Healthcheck (Ping to Redis)
	// This is going to verify the connectivity before accepting HTTP traffic
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
	} else {
		log.Println("Success: Connected to Redis")
	}

	// 4. Routes definition
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		// Simple example: increase a counter in Redis
		val, err := rdb.Incr(ctx, "access_count").Result()
		if err != nil {
			http.Error(w, "Error communicating with Redis", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Access count: %d\n", val)
	})

	// 5. Server Start
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
