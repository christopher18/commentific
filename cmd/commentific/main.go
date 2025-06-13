package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christopher18/commentific/v2/api"
	"github.com/christopher18/commentific/v2/models"
	"github.com/christopher18/commentific/v2/postgres"
	"github.com/christopher18/commentific/v2/service"
	"github.com/jmoiron/sqlx"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	Port        string
	Environment string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/commentific?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupDatabase initializes the database connection and runs migrations
func setupDatabase(databaseURL string) (*sqlx.DB, error) {
	// Open database connection
	db, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")

	// In a production application, you would run migrations here
	// For now, we'll just check if the tables exist
	if err := checkDatabaseSchema(db); err != nil {
		log.Printf("Warning: Database schema check failed: %v", err)
		log.Println("Please run the migration scripts in the migrations/ directory")
	}

	return db, nil
}

// checkDatabaseSchema performs a basic check to see if required tables exist
func checkDatabaseSchema(db *sqlx.DB) error {
	var exists bool

	// Check if comments table exists
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'comments'
		)`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check comments table: %w", err)
	}

	if !exists {
		return fmt.Errorf("comments table does not exist")
	}

	// Check if votes table exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'votes'
		)`).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check votes table: %w", err)
	}

	if !exists {
		return fmt.Errorf("votes table does not exist")
	}

	log.Println("Database schema check passed")
	return nil
}

// createCommentService creates and configures the comment service
func createCommentService(db *sqlx.DB) *service.CommentService {
	// Create repository provider
	provider := postgres.NewPostgresProvider(db)

	// Get comment repository
	repo := provider.GetCommentRepository()

	// Create service
	commentService := service.NewCommentService(repo)

	return commentService
}

// startServer starts the HTTP server
func startServer(config *Config, commentService *service.CommentService) *http.Server {
	// Create router
	router := api.NewRouter(commentService)

	// Create server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", config.Port)
		log.Printf("Environment: %s", config.Environment)
		log.Printf("API documentation available at: http://localhost:%s/", config.Port)
		log.Printf("Health check available at: http://localhost:%s/health", config.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	return server
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(server *http.Server, db *sqlx.DB) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Block until signal is received
	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server shutdown complete")
}

func main() {
	log.Println("Starting Commentific service...")

	// Load configuration
	config := loadConfig()

	// Setup database
	db, err := setupDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Create comment service
	commentService := createCommentService(db)

	// Start server
	server := startServer(config, commentService)

	// Wait for shutdown signal and handle graceful shutdown
	gracefulShutdown(server, db)
}

// Example usage patterns (these would typically be in separate example files)

// ExampleCreateComment demonstrates how to create a comment programmatically
func ExampleCreateComment() {
	// This is an example of how to use the comment service directly
	// without going through the HTTP API

	/*
		// Setup (you would do this once in your application)
		db, _ := sqlx.Open("postgres", "your-database-url")
		provider := postgres.NewPostgresProvider(db)
		repo := provider.GetCommentRepository()
		service := service.NewCommentService(repo)

		// Create a comment
		ctx := context.Background()
		req := &models.CreateCommentRequest{
			RootID:  "product-123",
			UserID:  "user-456",
			Content: "This is a great product!",
		}

		comment, err := service.CreateComment(ctx, req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Created comment: %+v\n", comment)

		// Create a reply
		replyReq := &models.CreateCommentRequest{
			RootID:   "product-123",
			ParentID: &comment.ID,
			UserID:   "user-789",
			Content:  "I agree, it's fantastic!",
		}

		reply, err := service.CreateComment(ctx, replyReq)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Created reply: %+v\n", reply)

		// Vote on the comment
		err = service.VoteComment(ctx, comment.ID, "user-789", models.VoteTypeUp)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Vote recorded successfully")
	*/
}

// ExampleGetCommentTree demonstrates how to retrieve a comment tree
func ExampleGetCommentTree() {
	/*
		// Get comment tree for a root
		tree, err := service.GetCommentTree(ctx, "product-123", 10, "score")
		if err != nil {
			log.Fatal(err)
		}

		// Print the tree structure
		for _, node := range tree {
			printNode(node, 0)
		}
	*/
}

// Helper function for printing tree structure
func printNode(node *models.CommentTree, depth int) {
	/*
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%s- %s (Score: %d)\n",
			indent, node.Comment.Content, node.Comment.Score)

		for _, child := range node.Children {
			printNode(child, depth+1)
		}
	*/
}
