package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/commentific/commentific/internal/models"
	"github.com/commentific/commentific/internal/repository/postgres"
	"github.com/commentific/commentific/internal/service"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Example of integrating Commentific into an existing application
// This shows how to use the service layer directly without the HTTP API

type ProductService struct {
	commentService *service.CommentService
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	// Initialize database (in real app, this would be shared)
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/commentific?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize commentific
	provider := postgres.NewPostgresProvider(db)
	repo := provider.GetCommentRepository()
	commentService := service.NewCommentService(repo)

	// Initialize your application service
	productService := &ProductService{
		commentService: commentService,
	}

	// Setup routes
	router := mux.NewRouter()

	// Your existing product routes
	router.HandleFunc("/products/{id}", productService.GetProduct).Methods("GET")

	// Integrated comment routes for products
	router.HandleFunc("/products/{id}/comments", productService.GetProductComments).Methods("GET")
	router.HandleFunc("/products/{id}/comments", productService.CreateProductComment).Methods("POST")
	router.HandleFunc("/products/{id}/comments/tree", productService.GetProductCommentTree).Methods("GET")

	fmt.Println("🚀 Integrated app with Commentific running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (ps *ProductService) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	// Mock product data
	product := &Product{
		ID:          productID,
		Name:        "Sample Product",
		Description: "This is a sample product with comments",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (ps *ProductService) GetProductComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	// Use commentific service to get comments for this product
	filter := &models.CommentFilter{
		Limit:  intPtr(20),
		Offset: intPtr(0),
	}

	comments, err := ps.commentService.GetCommentsByRoot(r.Context(), productID, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id": productID,
		"comments":   comments,
	})
}

func (ps *ProductService) CreateProductComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set the root ID to the product ID
	req.RootID = productID

	// In a real app, you'd get the user ID from authentication
	if req.UserID == "" {
		req.UserID = r.Header.Get("X-User-ID")
		if req.UserID == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}
	}

	comment, err := ps.commentService.CreateComment(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (ps *ProductService) GetProductCommentTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	maxDepth := 5 // Default depth
	if d := r.URL.Query().Get("max_depth"); d != "" {
		// Parse depth parameter (simplified)
		maxDepth = 5
	}

	tree, err := ps.commentService.GetCommentTree(r.Context(), productID, maxDepth, "score")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id":   productID,
		"comment_tree": tree,
		"max_depth":    maxDepth,
	})
}

func intPtr(i int) *int {
	return &i
}
