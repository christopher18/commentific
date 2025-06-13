package repository

import (
	"context"

	"github.com/christopher18/commentific/v2/models"
)

// CommentRepository defines the interface for comment data operations
type CommentRepository interface {
	// Comment CRUD operations
	CreateComment(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, id string) (*models.Comment, error)
	UpdateComment(ctx context.Context, id string, updates *models.UpdateCommentRequest) error
	DeleteComment(ctx context.Context, id string, userID string) error // Soft delete with user verification

	// Comment querying and filtering
	GetComments(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error)
	GetCommentsByRootID(ctx context.Context, rootID string, filter *models.CommentFilter) ([]*models.Comment, error)
	GetCommentsByUserID(ctx context.Context, userID string, filter *models.CommentFilter) ([]*models.Comment, error)
	GetCommentChildren(ctx context.Context, parentID string, maxDepth int) ([]*models.Comment, error)

	// Hierarchical operations
	GetCommentTree(ctx context.Context, rootID string, maxDepth int, sortBy string) ([]*models.CommentTree, error)
	GetCommentPath(ctx context.Context, commentID string) ([]*models.Comment, error) // Get path from root to comment

	// Vote operations
	CreateVote(ctx context.Context, vote *models.Vote) error
	UpdateVote(ctx context.Context, commentID, userID string, voteType models.VoteType) error
	DeleteVote(ctx context.Context, commentID, userID string) error
	GetUserVote(ctx context.Context, commentID, userID string) (*models.Vote, error)
	GetCommentVotes(ctx context.Context, commentID string) ([]*models.Vote, error)

	// Batch operations for performance
	GetCommentsWithUserVotes(ctx context.Context, rootID, userID string, filter *models.CommentFilter) ([]*models.Comment, map[string]*models.Vote, error)
	UpdateCommentScores(ctx context.Context, commentIDs []string) error

	// Statistics and analytics
	GetCommentStats(ctx context.Context, rootID string) (*models.CommentStats, error)
	GetUserCommentCount(ctx context.Context, userID string) (int64, error)
	GetTopComments(ctx context.Context, rootID string, limit int, timeRange string) ([]*models.Comment, error)

	// Maintenance operations
	PurgeDeletedComments(ctx context.Context, olderThan int) (int64, error) // Delete soft-deleted comments older than X days
	RecalculateCommentScores(ctx context.Context) error

	// Transaction support
	BeginTx(ctx context.Context) (Repository, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
}

// Repository interface for transaction support
type Repository interface {
	CommentRepository
}

// RepositoryProvider defines the interface for creating repository instances
type RepositoryProvider interface {
	GetCommentRepository() CommentRepository
	Close() error
	Health() error
	Migrate() error
}
