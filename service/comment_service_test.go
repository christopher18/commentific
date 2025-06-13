package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/christopher18/commentific/models"
	"github.com/christopher18/commentific/repository"
	"github.com/christopher18/commentific/service"
)

// MockRepository implements the CommentRepository interface for testing
type MockRepository struct {
	comments map[string]*models.Comment
	votes    map[string]*models.Vote
	error    error // Simulate repository errors
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		comments: make(map[string]*models.Comment),
		votes:    make(map[string]*models.Vote),
	}
}

// Implement CommentRepository interface methods (showing key ones)

func (m *MockRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	if m.error != nil {
		return m.error
	}

	// Simulate path calculation
	if comment.ParentID != nil {
		parent, exists := m.comments[*comment.ParentID]
		if !exists {
			return errors.New("parent comment not found")
		}
		comment.Depth = parent.Depth + 1
		comment.Path = parent.Path + "." + comment.ID
	} else {
		comment.Depth = 0
		comment.Path = comment.ID
	}

	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if m.error != nil {
		return nil, m.error
	}

	comment, exists := m.comments[id]
	if !exists {
		return nil, errors.New("comment not found")
	}
	return comment, nil
}

func (m *MockRepository) UpdateComment(ctx context.Context, id string, updates *models.UpdateCommentRequest) error {
	if m.error != nil {
		return m.error
	}

	comment, exists := m.comments[id]
	if !exists {
		return errors.New("comment not found")
	}

	if updates.Content != nil {
		comment.Content = *updates.Content
	}
	if updates.MediaURL != nil {
		comment.MediaURL = updates.MediaURL
	}
	if updates.LinkURL != nil {
		comment.LinkURL = updates.LinkURL
	}

	comment.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) DeleteComment(ctx context.Context, id string, userID string) error {
	if m.error != nil {
		return m.error
	}

	comment, exists := m.comments[id]
	if !exists {
		return errors.New("comment not found")
	}

	if comment.UserID != userID {
		return errors.New("user not authorized")
	}

	comment.IsDeleted = true
	comment.UpdatedAt = time.Now()
	return nil
}

func (m *MockRepository) UpdateVote(ctx context.Context, commentID, userID string, voteType models.VoteType) error {
	if m.error != nil {
		return m.error
	}

	// Check if comment exists
	comment, exists := m.comments[commentID]
	if !exists {
		return errors.New("comment not found")
	}

	// Create or update vote
	voteKey := commentID + ":" + userID
	vote := &models.Vote{
		ID:        voteKey,
		CommentID: commentID,
		UserID:    userID,
		VoteType:  voteType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.votes[voteKey] = vote

	// Update comment scores (simplified)
	upvotes := int64(0)
	downvotes := int64(0)
	for _, v := range m.votes {
		if v.CommentID == commentID {
			if v.VoteType == models.VoteTypeUp {
				upvotes++
			} else if v.VoteType == models.VoteTypeDown {
				downvotes++
			}
		}
	}

	comment.Upvotes = upvotes
	comment.Downvotes = downvotes
	comment.Score = upvotes - downvotes
	comment.UpdatedAt = time.Now()

	return nil
}

// Implement other required interface methods with minimal implementations
func (m *MockRepository) GetComments(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error) {
	if m.error != nil {
		return nil, m.error
	}

	var comments []*models.Comment
	for _, comment := range m.comments {
		if !comment.IsDeleted {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}

func (m *MockRepository) GetCommentsByRootID(ctx context.Context, rootID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if m.error != nil {
		return nil, m.error
	}

	var comments []*models.Comment
	for _, comment := range m.comments {
		if comment.RootID == rootID && !comment.IsDeleted {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}

func (m *MockRepository) GetCommentsByUserID(ctx context.Context, userID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if m.error != nil {
		return nil, m.error
	}

	var comments []*models.Comment
	for _, comment := range m.comments {
		if comment.UserID == userID && !comment.IsDeleted {
			comments = append(comments, comment)
		}
	}
	return comments, nil
}

// Add stub implementations for other interface methods to satisfy the interface
func (m *MockRepository) GetCommentChildren(ctx context.Context, parentID string, maxDepth int) ([]*models.Comment, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) GetCommentTree(ctx context.Context, rootID string, maxDepth int, sortBy string) ([]*models.CommentTree, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) GetCommentPath(ctx context.Context, commentID string) ([]*models.Comment, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) CreateVote(ctx context.Context, vote *models.Vote) error {
	return errors.New("not implemented in mock")
}

func (m *MockRepository) DeleteVote(ctx context.Context, commentID, userID string) error {
	return errors.New("not implemented in mock")
}

func (m *MockRepository) GetUserVote(ctx context.Context, commentID, userID string) (*models.Vote, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) GetCommentVotes(ctx context.Context, commentID string) ([]*models.Vote, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) GetCommentsWithUserVotes(ctx context.Context, rootID, userID string, filter *models.CommentFilter) ([]*models.Comment, map[string]*models.Vote, error) {
	return nil, nil, errors.New("not implemented in mock")
}

func (m *MockRepository) UpdateCommentScores(ctx context.Context, commentIDs []string) error {
	return errors.New("not implemented in mock")
}

func (m *MockRepository) GetCommentStats(ctx context.Context, rootID string) (*models.CommentStats, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) GetUserCommentCount(ctx context.Context, userID string) (int64, error) {
	return 0, errors.New("not implemented in mock")
}

func (m *MockRepository) GetTopComments(ctx context.Context, rootID string, limit int, timeRange string) ([]*models.Comment, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockRepository) PurgeDeletedComments(ctx context.Context, olderThan int) (int64, error) {
	return 0, errors.New("not implemented in mock")
}

func (m *MockRepository) RecalculateCommentScores(ctx context.Context) error {
	return errors.New("not implemented in mock")
}

func (m *MockRepository) BeginTx(ctx context.Context) (repository.Repository, error) {
	return m, nil
}

func (m *MockRepository) CommitTx(ctx context.Context) error {
	return nil
}

func (m *MockRepository) RollbackTx(ctx context.Context) error {
	return nil
}

// Test cases

func TestCreateComment_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Test data
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "This is a test comment",
	}

	// Execute
	comment, err := commentService.CreateComment(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if comment == nil {
		t.Fatal("Expected comment to be created, got nil")
	}

	if comment.RootID != req.RootID {
		t.Errorf("Expected root_id %s, got %s", req.RootID, comment.RootID)
	}

	if comment.UserID != req.UserID {
		t.Errorf("Expected user_id %s, got %s", req.UserID, comment.UserID)
	}

	if comment.Content != req.Content {
		t.Errorf("Expected content %s, got %s", req.Content, comment.Content)
	}

	if comment.Depth != 0 {
		t.Errorf("Expected depth 0 for root comment, got %d", comment.Depth)
	}

	if comment.ID == "" {
		t.Error("Expected comment ID to be generated")
	}
}

func TestCreateComment_InvalidContent(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Test data with empty content
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "",
	}

	// Execute
	comment, err := commentService.CreateComment(ctx, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty content, got nil")
	}

	if comment != nil {
		t.Error("Expected no comment to be created for invalid content")
	}
}

func TestCreateComment_WithParent(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create parent comment first
	parentReq := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Parent comment",
	}

	parent, err := commentService.CreateComment(ctx, parentReq)
	if err != nil {
		t.Fatalf("Failed to create parent comment: %v", err)
	}

	// Create child comment
	childReq := &models.CreateCommentRequest{
		RootID:   "test-root-1",
		ParentID: &parent.ID,
		UserID:   "user-456",
		Content:  "Child comment",
	}

	// Execute
	child, err := commentService.CreateComment(ctx, childReq)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if child.Depth != 1 {
		t.Errorf("Expected depth 1 for child comment, got %d", child.Depth)
	}

	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Error("Expected child to have correct parent ID")
	}

	expectedPath := parent.Path + "." + child.ID
	if child.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, child.Path)
	}
}

func TestVoteComment_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create a comment first
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Test comment",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Vote on the comment
	voterUserID := "user-456"
	err = commentService.VoteComment(ctx, comment.ID, voterUserID, models.VoteTypeUp)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify vote was recorded (check updated comment)
	updatedComment, err := commentService.GetComment(ctx, comment.ID)
	if err != nil {
		t.Fatalf("Failed to get updated comment: %v", err)
	}

	if updatedComment.Upvotes != 1 {
		t.Errorf("Expected 1 upvote, got %d", updatedComment.Upvotes)
	}

	if updatedComment.Score != 1 {
		t.Errorf("Expected score 1, got %d", updatedComment.Score)
	}
}

func TestVoteComment_SelfVote(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create a comment
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Test comment",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Try to vote on own comment
	err = commentService.VoteComment(ctx, comment.ID, comment.UserID, models.VoteTypeUp)

	// Assert
	if err == nil {
		t.Fatal("Expected error for self-vote, got nil")
	}
}

func TestUpdateComment_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create a comment
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Original content",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Update the comment
	newContent := "Updated content"
	updateReq := &models.UpdateCommentRequest{
		Content: &newContent,
	}

	err = commentService.UpdateComment(ctx, comment.ID, comment.UserID, updateReq)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify content was updated
	updatedComment, err := commentService.GetComment(ctx, comment.ID)
	if err != nil {
		t.Fatalf("Failed to get updated comment: %v", err)
	}

	if updatedComment.Content != newContent {
		t.Errorf("Expected content %s, got %s", newContent, updatedComment.Content)
	}
}

func TestUpdateComment_Unauthorized(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create a comment
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Original content",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Try to update with different user
	newContent := "Updated content"
	updateReq := &models.UpdateCommentRequest{
		Content: &newContent,
	}

	err = commentService.UpdateComment(ctx, comment.ID, "different-user", updateReq)

	// Assert
	if err == nil {
		t.Fatal("Expected error for unauthorized update, got nil")
	}
}

func TestDeleteComment_Success(t *testing.T) {
	// Setup
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Create a comment
	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Test comment",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Delete the comment
	err = commentService.DeleteComment(ctx, comment.ID, comment.UserID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify comment was soft deleted
	deletedComment, err := commentService.GetComment(ctx, comment.ID)
	if err != nil {
		t.Fatalf("Failed to get comment: %v", err)
	}

	if !deletedComment.IsDeleted {
		t.Error("Expected comment to be marked as deleted")
	}
}

// Benchmark tests

func BenchmarkCreateComment(b *testing.B) {
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &models.CreateCommentRequest{
			RootID:  "benchmark-root",
			UserID:  "benchmark-user",
			Content: "Benchmark comment content",
		}

		_, err := commentService.CreateComment(ctx, req)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkVoteComment(b *testing.B) {
	mockRepo := NewMockRepository()
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	// Setup: create a comment to vote on
	req := &models.CreateCommentRequest{
		RootID:  "benchmark-root",
		UserID:  "comment-author",
		Content: "Comment to vote on",
	}

	comment, err := commentService.CreateComment(ctx, req)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		voterID := "voter-" + string(rune(i))
		err := commentService.VoteComment(ctx, comment.ID, voterID, models.VoteTypeUp)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// Example test for testing with repository errors

func TestCreateComment_RepositoryError(t *testing.T) {
	// Setup mock with error
	mockRepo := NewMockRepository()
	mockRepo.error = errors.New("database connection failed")
	commentService := service.NewCommentService(mockRepo)
	ctx := context.Background()

	req := &models.CreateCommentRequest{
		RootID:  "test-root-1",
		UserID:  "user-123",
		Content: "Test comment",
	}

	// Execute
	comment, err := commentService.CreateComment(ctx, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error from repository, got nil")
	}

	if comment != nil {
		t.Error("Expected no comment when repository fails")
	}
}
