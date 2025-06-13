# Echo Integration Guide

This guide shows how to integrate Commentific into your existing Echo application.

## Quick Start

### 1. Install Dependencies

```bash
go get github.com/christopher18/commentific
go get github.com/labstack/echo/v4
```

### 2. Basic Integration

```go
package main

import (
    "github.com/christopher18/commentific/api"
    "github.com/christopher18/commentific/postgres"
    "github.com/christopher18/commentific/service"
    "github.com/labstack/echo/v4"
    "github.com/jmoiron/sqlx"
)

func main() {
    e := echo.New()

    // Initialize Commentific
    db, _ := sqlx.Connect("postgres", "your-db-url")
    provider := postgres.NewPostgresProvider(db)
    repo := provider.GetCommentRepository()
    commentService := service.NewCommentService(repo)

    // Register all Commentific routes
    commentAdapter := api.NewEchoAdapter(commentService)
    commentAdapter.RegisterRoutes(e)

    e.Start(":8080")
}
```

### 3. That's It!

Your Echo app now has all Commentific endpoints available:

- `POST /api/v1/comments` - Create comment
- `GET /api/v1/comments/:id` - Get comment
- `GET /api/v1/roots/:root_id/tree` - Get comment tree
- `GET /api/v1/comments/:id/children` - Get subtree
- And all other endpoints...

## Advanced Integration

### Custom Route Prefix

```go
// Register with custom prefix instead of /api/v1
commentAdapter.RegisterRoutesWithPrefix(e, "/comments/v1")
```

### Using Service Layer Directly

```go
// Use Commentific service in your own handlers
func getProductComments(commentService *service.CommentService) echo.HandlerFunc {
    return func(c echo.Context) error {
        productID := c.Param("id")
        
        tree, err := commentService.GetCommentTree(
            c.Request().Context(), 
            productID, 
            5, // max depth
            "score", // sort by
        )
        if err != nil {
            return c.JSON(500, map[string]string{"error": err.Error()})
        }
        
        return c.JSON(200, tree)
    }
}

// Register your custom handler
e.GET("/products/:id/comments", getProductComments(commentService))
```

### Middleware Integration

```go
// Add authentication middleware before Commentific routes
authGroup := e.Group("/api/v1")
authGroup.Use(yourAuthMiddleware)

// Register Commentific routes on the authenticated group
commentAdapter := api.NewEchoAdapter(commentService)
commentAdapter.RegisterRoutesWithPrefix(authGroup, "")
```

### Custom User ID Extraction

```go
// Middleware to extract user ID from JWT token
func extractUserID(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Extract user ID from JWT token
        userID := extractFromJWT(c.Request().Header.Get("Authorization"))
        
        // Set user ID header for Commentific
        c.Request().Header.Set("X-User-ID", userID)
        
        return next(c)
    }
}

// Apply to comment routes
commentGroup := e.Group("/api/v1")
commentGroup.Use(extractUserID)
commentAdapter.RegisterRoutesWithPrefix(commentGroup, "")
```

## Complete Example

See `examples/echo_integration/main.go` for a complete working example.

## API Endpoints

Once integrated, your Echo app will have these endpoints:

### Comment Operations
- `POST /api/v1/comments` - Create comment
- `GET /api/v1/comments/:id` - Get comment
- `PUT /api/v1/comments/:id` - Update comment
- `DELETE /api/v1/comments/:id` - Delete comment
- `GET /api/v1/comments/:id/children` - Get comment subtree
- `GET /api/v1/comments/:id/path` - Get comment path

### Voting Operations
- `POST /api/v1/comments/:id/vote` - Vote on comment
- `DELETE /api/v1/comments/:id/vote` - Remove vote

### Root-based Operations
- `GET /api/v1/roots/:root_id/comments` - Get comments for content
- `GET /api/v1/roots/:root_id/tree` - Get comment tree
- `GET /api/v1/roots/:root_id/stats` - Get statistics
- `GET /api/v1/roots/:root_id/top` - Get top comments
- `GET /api/v1/roots/:root_id/search` - Search comments

### User Operations
- `GET /api/v1/users/:user_id/comments` - Get user comments
- `GET /api/v1/users/:user_id/count` - Get user comment count

## Request/Response Format

### Create Comment
```bash
POST /api/v1/comments
Content-Type: application/json
X-User-ID: user123

{
  "root_id": "product-456",
  "parent_id": "comment-789", // optional for replies
  "content": "Great product!",
  "media_url": "https://example.com/image.jpg", // optional
  "link_url": "https://example.com/link" // optional
}
```

### Get Comment Tree
```bash
GET /api/v1/roots/product-456/tree?max_depth=5&sort_by=score
X-User-ID: user123
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "id": "comment-123",
      "content": "Great product!",
      "score": 15,
      "children": [
        {
          "id": "comment-456",
          "content": "I agree!",
          "score": 8,
          "children": []
        }
      ]
    }
  ]
}
```

## Error Handling

Commentific returns standard HTTP status codes:

- `200` - Success
- `201` - Created
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing user ID)
- `404` - Not Found
- `500` - Internal Server Error

Error response format:
```json
{
  "success": false,
  "error": "Error message here"
}
```

## Database Setup

Make sure to run the database migrations:

```bash
# Using golang-migrate
migrate -path migrations -database "postgres://..." up

# Or manually run the SQL files in migrations/
```

## Performance Tips

1. **Use pagination** for large comment threads
2. **Limit tree depth** to avoid deep recursion
3. **Cache frequently accessed trees** in Redis
4. **Use database connection pooling**
5. **Add indexes** for your specific query patterns

## Security Considerations

1. **Validate user permissions** before allowing operations
2. **Sanitize content** to prevent XSS
3. **Rate limit** comment creation
4. **Implement proper authentication**
5. **Use HTTPS** in production

## Next Steps

- Check out the [main README](../README.md) for more details
- See [examples/echo_integration/](../examples/echo_integration/) for complete code
- Read the [API documentation](../README.md#api-reference) for all endpoints 