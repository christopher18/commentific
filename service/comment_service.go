package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/christopher18/commentific/v2/models"
	"github.com/christopher18/commentific/v2/repository"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CommentService handles business logic for comments
type CommentService struct {
	repo      repository.CommentRepository
	validator *validator.Validate
}

// NewCommentService creates a new comment service
func NewCommentService(repo repository.CommentRepository) *CommentService {
	return &CommentService{
		repo:      repo,
		validator: validator.New(),
	}
}

// CreateComment creates a new comment with validation and business logic
func (s *CommentService) CreateComment(ctx context.Context, req *models.CreateCommentRequest) (*models.Comment, error) {
	// Validate the request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Sanitize content
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		return nil, fmt.Errorf("comment content cannot be empty")
	}

	// Validate URLs if provided
	if req.MediaURL != nil && *req.MediaURL != "" {
		if !s.isValidURL(*req.MediaURL) {
			return nil, fmt.Errorf("invalid media URL")
		}
	}

	if req.LinkURL != nil && *req.LinkURL != "" {
		if !s.isValidURL(*req.LinkURL) {
			return nil, fmt.Errorf("invalid link URL")
		}
	}

	// Create the comment model
	comment := &models.Comment{
		ID:       uuid.New().String(),
		RootID:   req.RootID,
		ParentID: req.ParentID,
		UserID:   req.UserID,
		Content:  req.Content,
		MediaURL: req.MediaURL,
		LinkURL:  req.LinkURL,
	}

	// Validate parent comment exists and belongs to same root if parentID is provided
	if req.ParentID != nil {
		parent, err := s.repo.GetCommentByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}
		if parent.RootID != req.RootID {
			return nil, fmt.Errorf("parent comment belongs to different root")
		}
		if parent.Depth >= 100 { // Prevent extremely deep nesting
			return nil, fmt.Errorf("maximum comment depth exceeded")
		}
	}

	// Create the comment
	err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetComment retrieves a comment by ID
func (s *CommentService) GetComment(ctx context.Context, id string) (*models.Comment, error) {
	if id == "" {
		return nil, fmt.Errorf("comment ID is required")
	}

	comment, err := s.repo.GetCommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

// UpdateComment updates a comment's content
func (s *CommentService) UpdateComment(ctx context.Context, id, userID string, req *models.UpdateCommentRequest) error {
	if id == "" {
		return fmt.Errorf("comment ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Get the existing comment to verify ownership
	comment, err := s.repo.GetCommentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	if comment.UserID != userID {
		return fmt.Errorf("user not authorized to update this comment")
	}

	// Validate and sanitize content if provided
	if req.Content != nil {
		*req.Content = strings.TrimSpace(*req.Content)
		if *req.Content == "" {
			return fmt.Errorf("comment content cannot be empty")
		}
		if len(*req.Content) > 10000 {
			return fmt.Errorf("comment content too long")
		}
	}

	// Validate URLs if provided
	if req.MediaURL != nil && *req.MediaURL != "" {
		if !s.isValidURL(*req.MediaURL) {
			return fmt.Errorf("invalid media URL")
		}
	}

	if req.LinkURL != nil && *req.LinkURL != "" {
		if !s.isValidURL(*req.LinkURL) {
			return fmt.Errorf("invalid link URL")
		}
	}

	return s.repo.UpdateComment(ctx, id, req)
}

// DeleteComment soft deletes a comment
func (s *CommentService) DeleteComment(ctx context.Context, id, userID string) error {
	if id == "" {
		return fmt.Errorf("comment ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	return s.repo.DeleteComment(ctx, id, userID)
}

// GetCommentsByRoot retrieves comments for a specific root with enhanced filtering
func (s *CommentService) GetCommentsByRoot(ctx context.Context, rootID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if rootID == "" {
		return nil, fmt.Errorf("root ID is required")
	}

	// Set default values for pagination
	if filter == nil {
		filter = &models.CommentFilter{}
	}

	// Set reasonable defaults
	if filter.Limit == nil {
		defaultLimit := 50
		filter.Limit = &defaultLimit
	}
	if filter.Offset == nil {
		defaultOffset := 0
		filter.Offset = &defaultOffset
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	// Enforce maximum limits to prevent abuse
	if *filter.Limit > 1000 {
		maxLimit := 1000
		filter.Limit = &maxLimit
	}

	return s.repo.GetCommentsByRootID(ctx, rootID, filter)
}

// GetCommentTree retrieves a hierarchical comment tree
func (s *CommentService) GetCommentTree(ctx context.Context, rootID string, maxDepth int, sortBy string) ([]*models.CommentTree, error) {
	if rootID == "" {
		return nil, fmt.Errorf("root ID is required")
	}

	// Set reasonable defaults
	if maxDepth <= 0 {
		maxDepth = 10 // Default max depth
	}
	if maxDepth > 50 {
		maxDepth = 50 // Prevent extremely deep trees
	}

	if sortBy == "" {
		sortBy = "score" // Default to sorting by score for tree view
	}

	return s.repo.GetCommentTree(ctx, rootID, maxDepth, sortBy)
}

// GetCommentsByUser retrieves comments by a specific user
func (s *CommentService) GetCommentsByUser(ctx context.Context, userID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Set default pagination
	if filter == nil {
		filter = &models.CommentFilter{}
	}
	if filter.Limit == nil {
		defaultLimit := 50
		filter.Limit = &defaultLimit
	}
	if filter.Offset == nil {
		defaultOffset := 0
		filter.Offset = &defaultOffset
	}

	return s.repo.GetCommentsByUserID(ctx, userID, filter)
}

// VoteComment handles voting on a comment
func (s *CommentService) VoteComment(ctx context.Context, commentID, userID string, voteType models.VoteType) error {
	if commentID == "" {
		return fmt.Errorf("comment ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if voteType != models.VoteTypeUp && voteType != models.VoteTypeDown {
		return fmt.Errorf("invalid vote type")
	}

	// Verify comment exists
	_, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	// Prevent users from voting on their own comments
	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}
	if comment.UserID == userID {
		return fmt.Errorf("users cannot vote on their own comments")
	}

	return s.repo.UpdateVote(ctx, commentID, userID, voteType)
}

// RemoveVote removes a user's vote from a comment
func (s *CommentService) RemoveVote(ctx context.Context, commentID, userID string) error {
	if commentID == "" {
		return fmt.Errorf("comment ID is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	return s.repo.DeleteVote(ctx, commentID, userID)
}

// GetCommentsWithUserVotes retrieves comments with user's voting status for efficient frontend rendering
func (s *CommentService) GetCommentsWithUserVotes(ctx context.Context, rootID, userID string, filter *models.CommentFilter) ([]*models.Comment, map[string]*models.Vote, error) {
	if rootID == "" {
		return nil, nil, fmt.Errorf("root ID is required")
	}
	if userID == "" {
		return nil, nil, fmt.Errorf("user ID is required")
	}

	// Set defaults
	if filter == nil {
		filter = &models.CommentFilter{}
	}
	if filter.Limit == nil {
		defaultLimit := 50
		filter.Limit = &defaultLimit
	}

	return s.repo.GetCommentsWithUserVotes(ctx, rootID, userID, filter)
}

// GetCommentStats retrieves statistics for a comment thread
func (s *CommentService) GetCommentStats(ctx context.Context, rootID string) (*models.CommentStats, error) {
	if rootID == "" {
		return nil, fmt.Errorf("root ID is required")
	}

	return s.repo.GetCommentStats(ctx, rootID)
}

// GetTopComments retrieves the highest-scored comments within a time range
func (s *CommentService) GetTopComments(ctx context.Context, rootID string, limit int, timeRange string) ([]*models.Comment, error) {
	if rootID == "" {
		return nil, fmt.Errorf("root ID is required")
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Prevent abuse
	}

	validTimeRanges := map[string]bool{
		"hour": true, "day": true, "week": true, "month": true, "all": true,
	}
	if !validTimeRanges[timeRange] {
		timeRange = "day" // Default to day
	}

	return s.repo.GetTopComments(ctx, rootID, limit, timeRange)
}

// GetUserCommentCount retrieves the total number of comments by a user
func (s *CommentService) GetUserCommentCount(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user ID is required")
	}

	return s.repo.GetUserCommentCount(ctx, userID)
}

// SearchComments searches for comments containing specific text
func (s *CommentService) SearchComments(ctx context.Context, rootID, query string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if rootID == "" {
		return nil, fmt.Errorf("root ID is required")
	}
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	query = strings.TrimSpace(query)
	if len(query) < 3 {
		return nil, fmt.Errorf("search query must be at least 3 characters")
	}

	// This is a simplified search - in production you might want to use
	// full-text search capabilities or external search services
	comments, err := s.repo.GetCommentsByRootID(ctx, rootID, filter)
	if err != nil {
		return nil, err
	}

	// Filter comments containing the query
	var results []*models.Comment
	queryLower := strings.ToLower(query)
	for _, comment := range comments {
		if strings.Contains(strings.ToLower(comment.Content), queryLower) {
			results = append(results, comment)
		}
	}

	return results, nil
}

// Maintenance Operations

// PurgeOldDeletedComments removes soft-deleted comments older than specified days
func (s *CommentService) PurgeOldDeletedComments(ctx context.Context, olderThanDays int) (int64, error) {
	if olderThanDays < 1 {
		return 0, fmt.Errorf("olderThanDays must be at least 1")
	}

	return s.repo.PurgeDeletedComments(ctx, olderThanDays)
}

// RecalculateAllScores recalculates vote scores for all comments
func (s *CommentService) RecalculateAllScores(ctx context.Context) error {
	return s.repo.RecalculateCommentScores(ctx)
}

// Utility methods

// isValidURL performs basic URL validation
func (s *CommentService) isValidURL(url string) bool {
	// Basic URL validation - in production, you might want more sophisticated validation
	url = strings.TrimSpace(url)
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// GetCommentPath retrieves the full path from root to a specific comment
func (s *CommentService) GetCommentPath(ctx context.Context, commentID string) ([]*models.Comment, error) {
	if commentID == "" {
		return nil, fmt.Errorf("comment ID is required")
	}

	return s.repo.GetCommentPath(ctx, commentID)
}

// GetCommentChildren retrieves all child comments for a given comment
func (s *CommentService) GetCommentChildren(ctx context.Context, parentID string, maxDepth int) ([]*models.Comment, error) {
	if parentID == "" {
		return nil, fmt.Errorf("parent ID is required")
	}

	if maxDepth <= 0 {
		maxDepth = 10
	}
	if maxDepth > 50 {
		maxDepth = 50
	}

	return s.repo.GetCommentChildren(ctx, parentID, maxDepth)
}

// BatchVoteComments allows voting on multiple comments at once (useful for bulk operations)
func (s *CommentService) BatchVoteComments(ctx context.Context, votes []models.VoteRequest, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if len(votes) > 100 {
		return fmt.Errorf("too many votes in batch, maximum is 100")
	}

	// Use transaction for batch operations
	repo, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			repo.RollbackTx(ctx)
		}
	}()

	for _, vote := range votes {
		// Basic validation
		if vote.UserID != userID {
			err = fmt.Errorf("user ID mismatch in vote request")
			return err
		}

		// Apply the vote
		err = repo.UpdateVote(ctx, "", vote.UserID, vote.VoteType)
		if err != nil {
			return fmt.Errorf("failed to apply vote: %w", err)
		}
	}

	return repo.CommitTx(ctx)
}

// CommentServiceConfig holds configuration for the comment service
type CommentServiceConfig struct {
	MaxCommentLength int
	MaxTreeDepth     int
	MaxBatchSize     int
	DefaultPageSize  int
	MaxPageSize      int
}

// NewCommentServiceWithConfig creates a comment service with custom configuration
func NewCommentServiceWithConfig(repo repository.CommentRepository, config *CommentServiceConfig) *CommentService {
	service := &CommentService{
		repo:      repo,
		validator: validator.New(),
	}

	// Apply configuration if provided
	if config != nil {
		// Configuration would be stored in service and used in validation
		// This is a placeholder for future extensibility
	}

	return service
}
