package api

import (
	"net/http"
	"time"

	"github.com/christopher18/commentific/service"
	"github.com/gorilla/mux"
)

// Router sets up and returns the HTTP router with all endpoints
func NewRouter(commentService *service.CommentService) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(corsMiddleware)
	router.Use(loggingMiddleware)
	router.Use(contentTypeMiddleware)

	// Create handler
	handler := NewCommentHandler(commentService)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Comment operations
	api.HandleFunc("/comments", handler.CreateComment).Methods("POST")
	api.HandleFunc("/comments/{id}", handler.GetComment).Methods("GET")
	api.HandleFunc("/comments/{id}", handler.UpdateComment).Methods("PUT")
	api.HandleFunc("/comments/{id}", handler.DeleteComment).Methods("DELETE")
	api.HandleFunc("/comments/{id}/path", handler.GetCommentPath).Methods("GET")
	api.HandleFunc("/comments/{id}/children", handler.GetCommentChildren).Methods("GET")

	// Voting operations
	api.HandleFunc("/comments/{id}/vote", handler.VoteComment).Methods("POST")
	api.HandleFunc("/comments/{id}/vote", handler.RemoveVote).Methods("DELETE")

	// Root-based operations (comments for specific entities)
	api.HandleFunc("/roots/{root_id}/comments", handler.GetCommentsByRoot).Methods("GET")
	api.HandleFunc("/roots/{root_id}/comments/with-votes", handler.GetCommentsWithVotes).Methods("GET")
	api.HandleFunc("/roots/{root_id}/tree", handler.GetCommentTree).Methods("GET")
	api.HandleFunc("/roots/{root_id}/stats", handler.GetCommentStats).Methods("GET")
	api.HandleFunc("/roots/{root_id}/top", handler.GetTopComments).Methods("GET")
	api.HandleFunc("/roots/{root_id}/search", handler.SearchComments).Methods("GET")

	// User operations
	api.HandleFunc("/users/{user_id}/comments", handler.GetCommentsByUser).Methods("GET")
	api.HandleFunc("/users/{user_id}/count", handler.GetUserCommentCount).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// API documentation endpoint
	router.HandleFunc("/", apiDocumentationHandler).Methods("GET")

	return router
}

// Middleware functions

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log request details (in production, use a proper logger)
		// fmt.Printf("[%s] %s %s - %d (%v)\n",
		//	time.Now().Format("2006-01-02 15:04:05"),
		//	r.Method, r.URL.Path, wrapped.statusCode, duration)

		// For now, we'll keep it simple to avoid import cycles
		_ = duration // Suppress unused variable warning
	})
}

// contentTypeMiddleware sets default content type for JSON responses
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "" &&
			(r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status": "healthy",
		"service": "commentific",
		"version": "1.0.0",
		"timestamp": "` + time.Now().Format(time.RFC3339) + `"
	}`))
}

// API documentation handler
func apiDocumentationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Commentific API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        h2 { color: #666; margin-top: 30px; }
        .endpoint { background-color: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #2e7d32; }
        .path { color: #1565c0; }
        code { background-color: #e8e8e8; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>Commentific API Documentation</h1>
    <p>A production-grade commenting system with infinite hierarchy support.</p>
    
    <h2>Authentication</h2>
    <p>Include user identification in requests using either:</p>
    <ul>
        <li>Header: <code>X-User-ID: your-user-id</code></li>
        <li>Query parameter: <code>?user_id=your-user-id</code></li>
    </ul>
    
    <h2>Comment Operations</h2>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/comments</span><br>
        Create a new comment
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/comments/{id}</span><br>
        Get a specific comment
    </div>
    
    <div class="endpoint">
        <span class="method">PUT</span> <span class="path">/api/v1/comments/{id}</span><br>
        Update a comment (requires ownership)
    </div>
    
    <div class="endpoint">
        <span class="method">DELETE</span> <span class="path">/api/v1/comments/{id}</span><br>
        Delete a comment (requires ownership)
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/comments/{id}/children</span><br>
        Get child comments (subtree) for a specific comment
        <br><small>Query params: <code>max_depth</code> (default: 10)</small>
    </div>
    
    <h2>Voting Operations</h2>
    
    <div class="endpoint">
        <span class="method">POST</span> <span class="path">/api/v1/comments/{id}/vote</span><br>
        Vote on a comment (body: {"vote_type": 1 or -1})
    </div>
    
    <div class="endpoint">
        <span class="method">DELETE</span> <span class="path">/api/v1/comments/{id}/vote</span><br>
        Remove vote from a comment
    </div>
    
    <h2>Root-based Operations</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/roots/{root_id}/comments</span><br>
        Get comments for a specific root (with pagination)
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/roots/{root_id}/tree</span><br>
        Get hierarchical comment tree
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/roots/{root_id}/stats</span><br>
        Get comment statistics for a root
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/roots/{root_id}/top</span><br>
        Get top-rated comments within time range
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/roots/{root_id}/search?q=query</span><br>
        Search comments within a root
    </div>
    
    <h2>User Operations</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/users/{user_id}/comments</span><br>
        Get comments by a specific user
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/api/v1/users/{user_id}/count</span><br>
        Get comment count for a user
    </div>
    
    <h2>Query Parameters</h2>
    <p>Most list endpoints support:</p>
    <ul>
        <li><code>limit</code> - Number of results (default: 50, max: 1000)</li>
        <li><code>offset</code> - Pagination offset</li>
        <li><code>sort_by</code> - Sort field (score, created_at, updated_at)</li>
        <li><code>sort_order</code> - Sort direction (asc, desc)</li>
        <li><code>max_depth</code> - Maximum comment depth for tree operations</li>
    </ul>
    
    <h2>Health Check</h2>
    <div class="endpoint">
        <span class="method">GET</span> <span class="path">/health</span><br>
        Service health status
    </div>
</body>
</html>`

	w.Write([]byte(html))
}
