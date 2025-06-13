# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2024-01-XX

### Fixed
- **Critical**: Fixed module name from `github.com/commentific/commentific` to `github.com/christopher18/commentific`
- Updated all import statements across the entire codebase
- Ensures proper Go module resolution and pkg.go.dev indexing

## [1.0.0] - 2024-01-XX

### Added
- Initial release of Commentific - Production-grade commenting system
- Infinite hierarchy threading with materialized path implementation
- Upvoting/downvoting system with automatic score calculation
- Media attachments and links support in comments
- Flexible root ID system for any content type (products, posts, articles, etc.)
- External user ID integration
- PostgreSQL database implementation with comprehensive indexing
- REST API with full CRUD operations
- Comment tree retrieval with configurable depth limits
- User-based comment queries and filtering
- Search functionality across comment content
- Statistics and analytics endpoints
- Batch operations for performance optimization
- Transaction support across all operations
- Database migrations (up/down) for schema management
- Comprehensive error handling and input validation
- CORS support and middleware
- Health check endpoints
- Built-in API documentation
- Clean architecture with dependency injection
- Extensible repository pattern for multiple database backends
- Production-ready logging and graceful shutdown
- Complete test suite with mocks and benchmarks

### Features
- **Hierarchical Comments**: Infinite nesting with efficient materialized path queries
- **Voting System**: Reddit-like upvote/downvote with score calculation
- **Media Support**: Attachments and link previews in comments
- **Flexible Integration**: Works with any content type via root IDs
- **Performance Optimized**: Database triggers, indexing, and batch operations
- **Production Ready**: Comprehensive error handling, validation, and monitoring
- **Extensible Design**: Interface-based architecture for easy customization

### API Endpoints
- `POST /api/v1/comments` - Create comment
- `GET /api/v1/comments/{id}` - Get comment
- `PUT /api/v1/comments/{id}` - Update comment
- `DELETE /api/v1/comments/{id}` - Delete comment
- `GET /api/v1/comments/{id}/children` - Get comment subtree
- `GET /api/v1/comments/{id}/path` - Get comment path from root
- `POST /api/v1/comments/{id}/vote` - Vote on comment
- `DELETE /api/v1/comments/{id}/vote` - Remove vote
- `GET /api/v1/roots/{root_id}/comments` - Get comments for content
- `GET /api/v1/roots/{root_id}/tree` - Get comment tree
- `GET /api/v1/roots/{root_id}/stats` - Get comment statistics
- `GET /api/v1/roots/{root_id}/top` - Get top comments
- `GET /api/v1/roots/{root_id}/search` - Search comments
- `GET /api/v1/users/{user_id}/comments` - Get user comments
- `GET /api/v1/users/{user_id}/count` - Get user comment count

### Database Schema
- Comments table with hierarchy support (path, depth, parent_id)
- Votes table with unique constraints per user/comment
- Comprehensive indexing for performance
- PostgreSQL triggers for automatic score updates
- Migration scripts for schema management

### Documentation
- Complete README with examples and integration guides
- API documentation with endpoint details
- Architecture documentation with design decisions
- Performance benchmarks and optimization guides
- Security considerations and best practices 