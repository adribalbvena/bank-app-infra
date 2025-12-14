package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func main() {
	// REDIS PASSWORD RETRIEVAL LOGIC
	// 1. Attempt to read from Vault injected file (Secure Production Path)
	redisPassword := ""
	// Vault injects the secret at this specific path
	vaultPath := "/vault/secrets/redis-config"

	content, err := os.ReadFile(vaultPath)
	if err == nil {
		// File exists, use its content as the password
		// TrimSpace is crucial to remove any newlines added by templates or editors
		redisPassword = strings.TrimSpace(string(content))
		log.Println("Config: Loaded Redis password from Vault file")
	} else {
		// 2. Fallback: Read from Environment Variable (Local Dev Path)
		// This ensures the app still works locally without Vault
		redisPassword = os.Getenv("REDIS_PASSWORD")
		if redisPassword != "" {
			log.Println("Config: Loaded Redis password from Env Var")
		} else {
			log.Println("Warning: No Redis password found (checked Vault file and ENV)")
		}
	}

	// REDIS CONNECTION
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local testing
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // Uses the resolved password
		DB:       0,             // Use default DB
	})

	// Verify Redis connection before starting the server
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
	} else {
		log.Println("Success: Connected to Redis")
	}

	// HTTP HANDLERS
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		// Increment a counter in Redis
		count, err := rdb.Incr(ctx, "access_count").Result()
		if err != nil {
			log.Printf("Error incrementing counter: %v", err)
			http.Error(w, "Error connecting to database", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Access count: %d", count)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Simple health check for Kubernetes liveness/readiness probes
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// SERVER START
	port := "8080"
	log.Printf("Server starting on port :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
