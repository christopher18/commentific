# Commentific Frontend API Specification

## Overview

Commentific is a production-grade commenting system that supports infinite hierarchical comment threading. This specification provides everything needed to build a frontend comment section that integrates with the Commentific API.

## Architecture & Design Patterns

### Hierarchical Comment Structure
- Uses **materialized path pattern** for storing comment hierarchies
- Root comments have path: `"uuid1"`
- Replies have path: `"uuid1.uuid2"` 
- Nested replies: `"uuid1.uuid2.uuid3"`
- Supports unlimited nesting depth

### Authentication Model
- **Bring Your Own Auth (BYOA)** - No built-in authentication
- User identification via headers or query parameters
- Client application handles authentication/authorization

### Key Design Patterns for Frontend

1. **Tree Rendering**: Use recursive components for nested comment display
2. **Optimistic Updates**: Update UI immediately, rollback on API errors
3. **Pagination**: Implement infinite scroll or pagination for large comment sets
4. **Real-time Feel**: Cache vote states and comment counts locally
5. **Progressive Enhancement**: Load comment trees incrementally by depth

## Base Configuration

**Base URL**: Your API server base URL
**API Version**: `/api/v1`
**Content-Type**: `application/json`

## Authentication & Headers

### Required Headers
```http
X-User-ID: string (required for write operations)
Content-Type: application/json (for POST/PUT requests)
```

### Alternative Authentication
```http
# Query parameter alternative
?user_id=your-user-id
```

## Data Models

### Comment Object
```typescript
interface Comment {
  id: string;                    // UUID
  root_id: string;              // External entity ID (article, post, etc.)
  parent_id?: string;           // Parent comment ID (null for root comments)
  user_id: string;              // External user ID
  content: string;              // Comment text content
  path: string;                 // Materialized path (e.g., "uuid1.uuid2")
  depth: number;                // Nesting level (0 = root)
  upvotes: number;              // Positive vote count
  downvotes: number;            // Negative vote count
  score: number;                // Calculated score (upvotes - downvotes)
  created_at: string;           // ISO 8601 timestamp
  updated_at: string;           // ISO 8601 timestamp
  is_edited: boolean;           // Whether comment was edited
  is_deleted: boolean;          // Soft delete flag
  reply_count: number;          // Number of direct replies
  total_replies: number;        // Total replies in subtree
}
```

### Vote Object
```typescript
interface Vote {
  comment_id: string;
  user_id: string;
  vote_type: 1 | -1;           // 1 for upvote, -1 for downvote
  created_at: string;
}
```

### Comment with Vote Status
```typescript
interface CommentWithVote extends Comment {
  user_vote?: {
    vote_type: 1 | -1;
    created_at: string;
  };
}
```

### API Response Wrappers
```typescript
interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
    has_more: boolean;
  };
}

interface ErrorResponse {
  error: string;
  details?: string;
  code?: string;
}
```

## API Endpoints

### Comment Operations

#### Create Comment
```http
POST /api/v1/comments
```

**Headers**: `X-User-ID: string`

**Body**:
```json
{
  "root_id": "article-123",
  "parent_id": "uuid-of-parent",  // Optional, null for root comments
  "content": "This is my comment content"
}
```

**Response**: `201 Created`
```json
{
  "id": "uuid-generated",
  "root_id": "article-123",
  "parent_id": "uuid-of-parent",
  "user_id": "user-123",
  "content": "This is my comment content",
  "path": "parent-uuid.new-uuid",
  "depth": 1,
  "upvotes": 0,
  "downvotes": 0,
  "score": 0,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "is_edited": false,
  "is_deleted": false,
  "reply_count": 0,
  "total_replies": 0
}
```

#### Get Comment
```http
GET /api/v1/comments/{id}
```

**Parameters**: `id` (path) - Comment UUID

**Query Parameters**:
- `user_id` (optional) - Include vote status for this user

**Response**: `200 OK` - Comment object

#### Update Comment
```http
PUT /api/v1/comments/{id}
```

**Headers**: `X-User-ID: string` (must match comment owner)

**Body**:
```json
{
  "content": "Updated comment content"
}
```

**Response**: `200 OK` - Updated Comment object

#### Delete Comment
```http
DELETE /api/v1/comments/{id}
```

**Headers**: `X-User-ID: string` (must match comment owner)

**Response**: `204 No Content`

#### Get Comment Path
```http
GET /api/v1/comments/{id}/path
```

**Response**: `200 OK`
```json
{
  "path": [
    {
      "id": "root-comment-id",
      "content": "Root comment...",
      "user_id": "user-1",
      "depth": 0
    },
    {
      "id": "parent-comment-id", 
      "content": "Parent comment...",
      "user_id": "user-2",
      "depth": 1
    }
  ]
}
```

#### Get Comment Children
```http
GET /api/v1/comments/{id}/children
```

**Query Parameters**:
- `max_depth` (optional, default: 10) - Maximum depth to retrieve
- `limit` (optional, default: 50) - Number of comments per level
- `user_id` (optional) - Include vote status

**Response**: `200 OK` - Array of Comment objects (hierarchically ordered)

### Voting Operations

#### Vote on Comment
```http
POST /api/v1/comments/{id}/vote
```

**Headers**: `X-User-ID: string`

**Body**:
```json
{
  "vote_type": 1  // 1 for upvote, -1 for downvote
}
```

**Response**: `200 OK`
```json
{
  "comment_id": "comment-uuid",
  "user_id": "user-123",
  "vote_type": 1,
  "created_at": "2024-01-01T12:00:00Z",
  "comment_score": 5  // Updated comment score
}
```

#### Remove Vote
```http
DELETE /api/v1/comments/{id}/vote
```

**Headers**: `X-User-ID: string`

**Response**: `204 No Content`

### Root-based Operations (Primary comment retrieval)

#### Get Comments by Root
```http
GET /api/v1/roots/{root_id}/comments
```

**Query Parameters**:
- `limit` (optional, default: 50, max: 1000) - Number of comments
- `offset` (optional, default: 0) - Pagination offset
- `sort_by` (optional, default: "score") - Sort field: "score", "created_at", "updated_at"
- `sort_order` (optional, default: "desc") - Sort direction: "asc", "desc"
- `user_id` (optional) - Include vote status for this user

**Response**: `200 OK` - PaginatedResponse<Comment>

#### Get Comments with Votes
```http
GET /api/v1/roots/{root_id}/comments/with-votes
```

**Headers**: `X-User-ID: string`

**Query Parameters**: Same as above

**Response**: `200 OK` - PaginatedResponse<CommentWithVote>

#### Get Comment Tree
```http
GET /api/v1/roots/{root_id}/tree
```

**Query Parameters**:
- `max_depth` (optional, default: 10) - Maximum nesting depth
- `limit` (optional, default: 50) - Comments per level
- `user_id` (optional) - Include vote status

**Response**: `200 OK`
```json
{
  "root_id": "article-123",
  "total_comments": 145,
  "tree": [
    {
      "id": "comment-1",
      "content": "Root comment",
      "depth": 0,
      "children": [
        {
          "id": "comment-2",
          "content": "Reply to root",
          "depth": 1,
          "children": []
        }
      ]
    }
  ]
}
```

#### Get Comment Stats
```http
GET /api/v1/roots/{root_id}/stats
```

**Response**: `200 OK`
```json
{
  "root_id": "article-123",
  "total_comments": 145,
  "root_comments": 23,
  "total_votes": 892,
  "unique_commenters": 67,
  "avg_score": 3.2,
  "last_comment_at": "2024-01-01T12:00:00Z"
}
```

#### Get Top Comments
```http
GET /api/v1/roots/{root_id}/top
```

**Query Parameters**:
- `limit` (optional, default: 10) - Number of top comments
- `time_range` (optional, default: "all") - "hour", "day", "week", "month", "all"
- `min_score` (optional, default: 1) - Minimum score threshold

**Response**: `200 OK` - Array of Comment objects sorted by score

#### Search Comments
```http
GET /api/v1/roots/{root_id}/search
```

**Query Parameters**:
- `q` (required) - Search query
- `limit` (optional, default: 50) - Number of results
- `offset` (optional, default: 0) - Pagination offset

**Response**: `200 OK` - PaginatedResponse<Comment>

### User Operations

#### Get User Comments
```http
GET /api/v1/users/{user_id}/comments
```

**Query Parameters**:
- `limit` (optional, default: 50) - Number of comments
- `offset` (optional, default: 0) - Pagination offset
- `sort_by` (optional, default: "created_at") - Sort field
- `sort_order` (optional, default: "desc") - Sort direction

**Response**: `200 OK` - PaginatedResponse<Comment>

#### Get User Comment Count
```http
GET /api/v1/users/{user_id}/count
```

**Response**: `200 OK`
```json
{
  "user_id": "user-123",
  "total_comments": 42,
  "total_votes_received": 156
}
```

### Health Check

#### Service Health
```http
GET /health
```

**Response**: `200 OK`
```json
{
  "status": "healthy",
  "service": "commentific",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Human readable error message",
  "details": "Additional context or validation errors",
  "code": "ERROR_CODE"
}
```

### Common HTTP Status Codes

- `200 OK` - Success
- `201 Created` - Resource created successfully
- `204 No Content` - Success with no response body
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate vote)
- `422 Unprocessable Entity` - Validation errors
- `500 Internal Server Error` - Server error

## Frontend Implementation Patterns

### 1. Comment Tree Rendering

```typescript
// Recursive component pattern
const CommentTree = ({ comments, depth = 0, maxDepth = 10 }) => {
  return (
    <div className={`comment-level-${depth}`}>
      {comments.map(comment => (
        <CommentItem 
          key={comment.id} 
          comment={comment} 
          depth={depth}
        >
          {comment.children && depth < maxDepth && (
            <CommentTree 
              comments={comment.children}
              depth={depth + 1}
              maxDepth={maxDepth}
            />
          )}
        </CommentItem>
      ))}
    </div>
  );
};
```

### 2. Optimistic Updates

```typescript
const handleVote = async (commentId: string, voteType: 1 | -1) => {
  // Optimistic update
  updateCommentScoreLocally(commentId, voteType);
  
  try {
    await api.post(`/comments/${commentId}/vote`, { vote_type: voteType });
  } catch (error) {
    // Rollback on error
    revertCommentScoreLocally(commentId, voteType);
    showError('Failed to vote');
  }
};
```

### 3. Infinite Scroll Implementation

```typescript
const useInfiniteComments = (rootId: string) => {
  const [comments, setComments] = useState([]);
  const [hasMore, setHasMore] = useState(true);
  const [offset, setOffset] = useState(0);
  
  const loadMore = async () => {
    const response = await api.get(`/roots/${rootId}/comments`, {
      params: { limit: 20, offset }
    });
    
    setComments(prev => [...prev, ...response.data.data]);
    setHasMore(response.data.pagination.has_more);
    setOffset(prev => prev + 20);
  };
  
  return { comments, loadMore, hasMore };
};
```

### 4. Real-time Updates

```typescript
// Poll for new comments
const useCommentPolling = (rootId: string, interval = 30000) => {
  const [lastUpdate, setLastUpdate] = useState(new Date());
  
  useEffect(() => {
    const pollForUpdates = async () => {
      const response = await api.get(`/roots/${rootId}/comments`, {
        params: { 
          since: lastUpdate.toISOString(),
          limit: 50 
        }
      });
      
      if (response.data.data.length > 0) {
        updateCommentsLocally(response.data.data);
        setLastUpdate(new Date());
      }
    };
    
    const timer = setInterval(pollForUpdates, interval);
    return () => clearInterval(timer);
  }, [rootId, lastUpdate, interval]);
};
```

### 5. Form Handling

```typescript
const CommentForm = ({ parentId, rootId, onSubmit }) => {
  const [content, setContent] = useState('');
  const [loading, setLoading] = useState(false);
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const response = await api.post('/comments', {
        root_id: rootId,
        parent_id: parentId,
        content
      });
      
      onSubmit(response.data);
      setContent('');
    } catch (error) {
      showError('Failed to post comment');
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <textarea 
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Write a comment..."
        maxLength={10000}
        required
      />
      <button type="submit" disabled={loading || !content.trim()}>
        {loading ? 'Posting...' : 'Post Comment'}
      </button>
    </form>
  );
};
```

## Best Practices

### Performance
- Use pagination for large comment sets
- Implement lazy loading for deep comment trees
- Cache comment data locally
- Use debounced search queries

### User Experience  
- Show loading states during API calls
- Implement optimistic updates for votes
- Provide clear feedback for errors
- Support keyboard navigation

### Security
- Sanitize user input before display
- Validate comment content length
- Implement rate limiting on client side
- Never trust client-side data

### Accessibility
- Use semantic HTML structure
- Implement proper ARIA labels
- Support keyboard navigation
- Ensure color contrast compliance

## Rate Limiting Recommendations

- Comment creation: 1 per 10 seconds per user
- Voting: 10 per minute per user  
- Comment updates: 1 per 30 seconds per user
- Search queries: 30 per minute per user

## Caching Strategy --> DO NOT IMPLEMENT YET!!

- Cache comment trees for 5 minutes --> DO NOT IMPLEMENT YET!!
- Cache user vote status for 10 minutes --> DO NOT IMPLEMENT YET!!
- Cache comment counts for 2 minutes --> DO NOT IMPLEMENT YET!!
- Invalidate cache on user actions --> DO NOT IMPLEMENT YET!!

This specification provides everything needed to build a robust, scalable comment section that integrates seamlessly with the Commentific API. 