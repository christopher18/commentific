# Commentific

A production-grade, horizontally-scalable commenting system for any application. Commentific provides Reddit-like features including infinite hierarchy threading, voting, media attachments, and comprehensive search capabilities.

## ✨ Features

- **Infinite Hierarchy Threading** - Nested comments with unlimited depth
- **Voting System** - Upvote/downvote with automatic score calculation
- **Edit Tracking** - Comprehensive edit history with original content preservation
- **Media Support** - Attach images, videos, and links to comments
- **Search & Filtering** - Full-text search with advanced filtering options
- **Database Agnostic** - PostgreSQL implementation with extensible interface for other databases
- **REST API** - Complete HTTP API with OpenAPI documentation
- **Go Module** - Use as a library in your Go applications
- **Production Ready** - Optimized queries, connection pooling, graceful shutdown
- **Flexible Root IDs** - Comments can be attached to any entity (posts, products, articles, etc.)
- **External User IDs** - Integrate with your existing user system

## 🚀 Quick Start

### Prerequisites (Required for ALL usage methods)

1. **Set up PostgreSQL database:**
```sql
CREATE DATABASE commentific;
CREATE USER commentific WITH PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE commentific TO commentific;
```

2. **Install required PostgreSQL extension:**
```sql
-- Connect as superuser (postgres) to install extension
\c commentific
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

3. **Run database migrations:**
```bash
# Apply the migrations to create tables and indexes
psql -d commentific -f migrations/001_create_comments_table.up.sql
psql -d commentific -f migrations/002_add_edit_tracking.up.sql
```

### Option 1: As a Standalone Service

4. **Set environment variables:**
```bash
export DATABASE_URL="postgres://commentific:your-password@localhost/commentific?sslmode=disable"
export PORT="8080"
export ENVIRONMENT="production"
```

5. **Run the service:**
```bash
go run cmd/commentific/main.go
```

6. **Visit the API documentation:**
Open http://localhost:8080/ in your browser to see the interactive API documentation.

### Option 2: As a Go Module

4. **Install the module:**
```bash
go get github.com/christopher18/commentific
```

**Framework Integration:**
- 🚀 **[Echo Framework](docs/ECHO_INTEGRATION.md)** - Complete integration guide with one-line setup
- 📝 **Gin Framework** - See integration examples below

5. **Use in your Go application:**

```go
package main

import (
    "context"
    "log"
    
    "github.com/christopher18/commentific/models"
    "github.com/christopher18/commentific/postgres"
    "github.com/christopher18/commentific/service"
    "github.com/jmoiron/sqlx"
)

func main() {
    // Connect to the SAME database where you ran the migrations
    db, err := sqlx.Connect("postgres", "postgres://commentific:your-password@localhost/commentific?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create the comment service
    provider := postgres.NewPostgresProvider(db)
    repo := provider.GetCommentRepository()
    commentService := service.NewCommentService(repo)
    
    // Create a comment
    ctx := context.Background()
    req := &models.CreateCommentRequest{
        RootID:  "product-123",
        UserID:  "user-456", 
        Content: "This is a great product!",
    }
    
    comment, err := commentService.CreateComment(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created comment: %+v", comment)
}
```

**Important:** Both usage methods require the same database setup and migrations. The difference is that as a Go module, YOU manage the database connection and HTTP routing.

## 📖 API Reference

### Authentication

Include user identification in requests using either:
- Header: `X-User-ID: your-user-id`
- Query parameter: `?user_id=your-user-id`

### Core Operations

#### Create Comment
```http
POST /api/v1/comments
Content-Type: application/json

{
  "root_id": "product-123",
  "parent_id": "optional-parent-comment-id",
  "user_id": "user-456",
  "content": "This is my comment",
  "media_url": "https://example.com/image.jpg",
  "link_url": "https://example.com/related-link"
}
```

#### Get Comment Tree
```http
GET /api/v1/roots/product-123/tree?max_depth=10&sort_by=score
```

#### Vote on Comment
```http
POST /api/v1/comments/{comment-id}/vote
Content-Type: application/json

{
  "vote_type": 1  // 1 for upvote, -1 for downvote
}
```

#### Search Comments
```http
GET /api/v1/roots/product-123/search?q=searchterm&limit=20
```

#### Update Comment
```http
PUT /api/v1/comments/{comment-id}
Content-Type: application/json
X-User-ID: user-456

{
  "content": "Updated comment text"
}
```

#### Get Edited Comments
```http
GET /api/v1/roots/product-123/edited?min_edits=2&sort_by=edit_count
```

### Edit Tracking

Commentific automatically tracks all comment edits with comprehensive metadata:

#### Edit Information Included
All comment responses include edit tracking fields:
```json
{
  "id": "comment-uuid",
  "content": "Updated comment text",
  "is_edited": true,
  "edit_count": 3,
  "original_content": "Original comment text",
  "content_updated_at": "2024-01-15T14:30:22Z",
  "created_at": "2024-01-10T10:15:30Z",
  "updated_at": "2024-01-15T14:30:22Z"
}
```

#### Edit Filtering & Querying
```http
# Get only edited comments
GET /api/v1/roots/product-123/comments?is_edited=true

# Get comments with at least 2 edits
GET /api/v1/roots/product-123/comments?min_edits=2

# Sort by most edited first
GET /api/v1/roots/product-123/comments?sort_by=edit_count&sort_order=desc

# Sort by most recently edited
GET /api/v1/roots/product-123/comments?sort_by=content_updated_at&sort_order=desc

# Get dedicated edited comments endpoint
GET /api/v1/roots/product-123/edited
```

#### How Edit Tracking Works
- **Automatic Detection**: Database triggers detect content changes
- **Original Preservation**: First edit preserves original content
- **Edit Counting**: Tracks total number of modifications
- **Smart Triggers**: Only content changes trigger edit tracking (not votes)
- **Zero Overhead**: No application logic required

### Response Format

All API responses follow this format:
```json
{
  "success": true,
  "data": { ... },
  "message": "Optional message",
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 1234
  }
}
```

Error responses:
```json
{
  "success": false,
  "error": "Error description"
}
```

## 🏗 Architecture

Commentific follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────────┐
│   HTTP API      │  (REST endpoints, JSON serialization)
├─────────────────┤
│  Service Layer  │  (Business logic, validation)
├─────────────────┤
│ Repository Layer│  (Data access interface)
├─────────────────┤
│   PostgreSQL    │  (Database implementation)
└─────────────────┘
```

### Key Design Decisions

- **Materialized Path**: Comments use a materialized path pattern for efficient tree queries
- **Separate Vote Table**: Votes are stored separately to enable complex voting logic
- **External IDs**: Both user IDs and root IDs are external to support any application
- **Soft Deletes**: Comments are soft-deleted to maintain thread integrity
- **Optimistic Scoring**: Vote scores are calculated via database triggers for consistency

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://user:password@localhost/commentific?sslmode=disable` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `ENVIRONMENT` | `development` | Environment (development/production) |

### Database Configuration

**Required Extension:**
```sql
-- Required for text search functionality (must be installed by superuser)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

For production deployments, consider these PostgreSQL settings:

```sql
-- Connection pooling
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
```

**Note:** If you get an error about `gist_trgm_ops` not existing, you need to install the `pg_trgm` extension as shown above.

## 🔌 Integration Examples

### With Gin Web Framework

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/christopher18/commentific/service"
)

func setupCommentRoutes(r *gin.Engine, commentService *service.CommentService) {
    api := r.Group("/api/v1")
    
    api.POST("/comments", func(c *gin.Context) {
        var req models.CreateCommentRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // Get user ID from your auth middleware
        req.UserID = c.GetString("user_id")
        
        comment, err := commentService.CreateComment(c.Request.Context(), &req)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(201, comment)
    })
}
```

### With Echo Framework

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/christopher18/commentific/api"
    "github.com/christopher18/commentific/service"
)

func main() {
    e := echo.New()
    
    // Initialize Commentific
    commentService := service.NewCommentService(repo)
    
    // Option 1: Register all routes automatically
    commentAdapter := api.NewEchoAdapter(commentService)
    commentAdapter.RegisterRoutes(e)
    
    // Option 2: Use service in your own handlers
    e.GET("/products/:id/comments", getProductComments(commentService))
    
    e.Start(":8080")
}

func getProductComments(commentService *service.CommentService) echo.HandlerFunc {
    return func(c echo.Context) error {
        productID := c.Param("id")
        tree, err := commentService.GetCommentTree(c.Request().Context(), productID, 5, "score")
        if err != nil {
            return echo.NewHTTPError(500, err.Error())
        }
        return c.JSON(200, tree)
    }
}
```

📖 **[Complete Echo Integration Guide](docs/ECHO_INTEGRATION.md)** - Detailed documentation with advanced examples, middleware integration, and best practices.

## 🧪 Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Performance

### Benchmarks

- **Comment Creation**: ~1,000 ops/sec
- **Tree Retrieval**: ~5,000 ops/sec for 1000-comment trees
- **Vote Recording**: ~2,000 ops/sec
- **Search**: ~500 ops/sec for full-text search

### Scaling Recommendations

- **Read Replicas**: Use PostgreSQL read replicas for read-heavy workloads
- **Caching**: Implement Redis caching for frequently accessed comment trees
- **Database Partitioning**: Partition by root_id for very large datasets
- **CDN**: Use CDN for media URLs to reduce server load

## 🔒 Security Considerations

- **SQL Injection**: All queries use parameterized statements
- **Input Validation**: Comprehensive validation on all inputs
- **Rate Limiting**: Implement rate limiting in your application layer
- **Authentication**: Bring your own authentication system
- **Content Moderation**: Implement content filtering as needed

## 🛠 Development

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Make (optional)

### Setup Development Environment

1. Clone the repository:
```bash
git clone https://github.com/christopher18/commentific.git
cd commentific
```

2. Start PostgreSQL (using Docker):
```bash
docker run --name commentific-postgres \
  -e POSTGRES_DB=commentific \
  -e POSTGRES_USER=commentific \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -d postgres:13
```

3. Run migrations:
```bash
psql -h localhost -U commentific -d commentific -f migrations/001_create_comments_table.up.sql
```

4. Run the application:
```bash
go run cmd/commentific/main.go
```

### Adding New Database Backends

To add support for a new database (e.g., MySQL, SQLite):

1. Create a new package under `internal/repository/` (e.g., `mysql`)
2. Implement the `CommentRepository` interface
3. Create a provider that implements `RepositoryProvider`
4. Update the main application to support the new backend

Example structure:
```
internal/repository/mysql/
├── mysql.go           # Repository implementation
├── queries.go         # SQL queries
└── migrations/        # MySQL-specific migrations
```

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

- 📚 [Documentation](https://github.com/christopher18/commentific/wiki)
- 🐛 [Issue Tracker](https://github.com/christopher18/commentific/issues)
- 💬 [Discussions](https://github.com/christopher18/commentific/discussions)

## 🗺 Roadmap

- [ ] Redis caching layer
- [ ] Elasticsearch integration for advanced search
- [ ] WebSocket support for real-time comments
- [ ] Content moderation hooks
- [ ] MySQL and SQLite support
- [ ] GraphQL API
- [ ] Comment reactions (beyond up/down votes)
- [ ] Media processing pipeline
- [ ] Analytics and reporting features

---

Made with ❤️ for the developer community 