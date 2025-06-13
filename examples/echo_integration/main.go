package main

import (
	"log"
	"net/http"

	"github.com/christopher18/commentific/api"
	"github.com/christopher18/commentific/postgres"
	"github.com/christopher18/commentific/service"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize your Echo app
	e := echo.New()

	// Add your existing middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Your existing routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to my app with Commentific!",
		})
	})

	// Initialize Commentific
	db, err := sqlx.Connect("postgres", "postgres://user:password@localhost/commentific?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	provider := postgres.NewPostgresProvider(db)
	repo := provider.GetCommentRepository()
	commentService := service.NewCommentService(repo)

	// Create Echo adapter and register all Commentific routes
	commentAdapter := api.NewEchoAdapter(commentService)

	// Option 1: Register with default prefix (/api/v1)
	commentAdapter.RegisterRoutes(e)

	// Option 2: Register with custom prefix (uncomment to use instead)
	// commentAdapter.RegisterRoutesWithPrefix(e, "/comments/v1")

	// Your app-specific routes that use comments
	e.GET("/products/:id", getProduct)
	e.GET("/products/:id/comments", getProductComments(commentService))

	log.Println("🚀 Echo app with Commentific running on :8080")
	log.Println("📖 Available endpoints:")
	log.Println("   GET  /                              - App home")
	log.Println("   GET  /products/:id                  - Get product")
	log.Println("   GET  /products/:id/comments         - Get product comments")
	log.Println("   POST /api/v1/comments               - Create comment")
	log.Println("   GET  /api/v1/roots/:root_id/tree    - Get comment tree")
	log.Println("   ... and all other Commentific endpoints")

	e.Logger.Fatal(e.Start(":8080"))
}

// Your existing handlers
func getProduct(c echo.Context) error {
	productID := c.Param("id")

	// Your product logic here
	product := map[string]interface{}{
		"id":          productID,
		"name":        "Sample Product",
		"description": "This product has comments powered by Commentific",
	}

	return c.JSON(http.StatusOK, product)
}

// Example of using Commentific service directly in your handlers
func getProductComments(commentService *service.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		productID := c.Param("id")

		// Use Commentific service to get comments for this product
		tree, err := commentService.GetCommentTree(c.Request().Context(), productID, 5, "score")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"product_id":   productID,
			"comment_tree": tree,
		})
	}
}
