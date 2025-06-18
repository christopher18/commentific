package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/christopher18/commentific/v2/models"
	"github.com/christopher18/commentific/v2/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PostgresRepository implements the CommentRepository interface for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

// PostgresProvider implements the RepositoryProvider interface
type PostgresProvider struct {
	db *sqlx.DB
}

// NewPostgresProvider creates a new PostgreSQL repository provider
func NewPostgresProvider(db *sqlx.DB) *PostgresProvider {
	return &PostgresProvider{db: db}
}

// GetCommentRepository returns a PostgreSQL comment repository
func (p *PostgresProvider) GetCommentRepository() repository.CommentRepository {
	return &PostgresRepository{db: p.db}
}

// Close closes the database connection
func (p *PostgresProvider) Close() error {
	return p.db.Close()
}

// Health checks if the database is healthy
func (p *PostgresProvider) Health() error {
	return p.db.Ping()
}

// Migrate runs database migrations
func (p *PostgresProvider) Migrate() error {
	// This would integrate with a migration tool like golang-migrate
	// For now, we'll implement a simple check
	var exists bool
	err := p.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'comments')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}

	if !exists {
		return fmt.Errorf("database not migrated: comments table does not exist")
	}

	return nil
}

// getDB returns the appropriate database connection (transaction or regular)
func (r *PostgresRepository) getDB() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// getQueryable returns a queryable interface that supports Get and Select methods
func (r *PostgresRepository) getQueryable() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// CreateComment creates a new comment
func (r *PostgresRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	// Generate ID if not provided
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}

	// Calculate path and depth
	if comment.ParentID != nil {
		// Get parent comment to calculate path and depth
		parent, err := r.GetCommentByID(ctx, *comment.ParentID)
		if err != nil {
			return fmt.Errorf("failed to get parent comment: %w", err)
		}
		comment.Depth = parent.Depth + 1
		comment.Path = parent.Path + "." + comment.ID

		// Verify parent belongs to same root
		if parent.RootID != comment.RootID {
			return fmt.Errorf("parent comment belongs to different root")
		}
	} else {
		comment.Depth = 0
		comment.Path = comment.ID
	}

	query := `
		INSERT INTO comments (id, root_id, parent_id, user_id, content, media_url, link_url, depth, path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	_, err := r.getDB().ExecContext(ctx, query,
		comment.ID, comment.RootID, comment.ParentID, comment.UserID,
		comment.Content, comment.MediaURL, comment.LinkURL, comment.Depth,
		comment.Path, comment.CreatedAt, comment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetCommentByID retrieves a comment by its ID
func (r *PostgresRepository) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	query := `
		SELECT id, root_id, parent_id, user_id, content, media_url, link_url, 
		       upvotes, downvotes, score, depth, path, is_deleted, created_at, updated_at
		FROM comments 
		WHERE id = $1 AND NOT is_deleted`

	comment := &models.Comment{}
	err := r.getQueryable().GetContext(ctx, comment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("comment not found")
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

// UpdateComment updates a comment's content
func (r *PostgresRepository) UpdateComment(ctx context.Context, id string, updates *models.UpdateCommentRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *updates.Content)
		argIndex++
	}

	if updates.MediaURL != nil {
		setParts = append(setParts, fmt.Sprintf("media_url = $%d", argIndex))
		args = append(args, *updates.MediaURL)
		argIndex++
	}

	if updates.LinkURL != nil {
		setParts = append(setParts, fmt.Sprintf("link_url = $%d", argIndex))
		args = append(args, *updates.LinkURL)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil // Nothing to update
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf("UPDATE comments SET %s WHERE id = $%d AND NOT is_deleted",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	result, err := r.getDB().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found or already deleted")
	}

	return nil
}

// DeleteComment soft deletes a comment
func (r *PostgresRepository) DeleteComment(ctx context.Context, id string, userID string) error {
	query := `UPDATE comments SET is_deleted = true, updated_at = $1 WHERE id = $2 AND user_id = $3 AND NOT is_deleted`

	result, err := r.getDB().ExecContext(ctx, query, time.Now(), id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("comment not found, already deleted, or user not authorized")
	}

	return nil
}

// GetComments retrieves comments based on filter
func (r *PostgresRepository) GetComments(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error) {
	query := `
		SELECT id, root_id, parent_id, user_id, content, media_url, link_url,
		       upvotes, downvotes, score, depth, path, is_deleted, created_at, updated_at
		FROM comments 
		WHERE NOT is_deleted`

	args := []interface{}{}
	argIndex := 1

	if filter.RootID != nil {
		query += fmt.Sprintf(" AND root_id = $%d", argIndex)
		args = append(args, *filter.RootID)
		argIndex++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.ParentID != nil {
		query += fmt.Sprintf(" AND parent_id = $%d", argIndex)
		args = append(args, *filter.ParentID)
		argIndex++
	}

	if filter.MaxDepth != nil {
		query += fmt.Sprintf(" AND depth <= $%d", argIndex)
		args = append(args, *filter.MaxDepth)
		argIndex++
	}

	// Add sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "score", "created_at", "updated_at":
			sortBy = filter.SortBy
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	if filter.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *filter.Limit)
		argIndex++
	}

	if filter.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, *filter.Offset)
		argIndex++
	}

	comments := []*models.Comment{}
	err := r.getQueryable().SelectContext(ctx, &comments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return comments, nil
}

// GetCommentsByRootID retrieves comments for a specific root
func (r *PostgresRepository) GetCommentsByRootID(ctx context.Context, rootID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if filter == nil {
		filter = &models.CommentFilter{}
	}
	filter.RootID = &rootID
	return r.GetComments(ctx, filter)
}

// GetCommentsByUserID retrieves comments by a specific user
func (r *PostgresRepository) GetCommentsByUserID(ctx context.Context, userID string, filter *models.CommentFilter) ([]*models.Comment, error) {
	if filter == nil {
		filter = &models.CommentFilter{}
	}
	filter.UserID = &userID
	return r.GetComments(ctx, filter)
}

// GetCommentChildren retrieves child comments up to maxDepth
func (r *PostgresRepository) GetCommentChildren(ctx context.Context, parentID string, maxDepth int) ([]*models.Comment, error) {
	query := `
		SELECT id, root_id, parent_id, user_id, content, media_url, link_url,
		       upvotes, downvotes, score, depth, path, is_deleted, created_at, updated_at
		FROM comments 
		WHERE path LIKE $1 AND NOT is_deleted AND depth <= $2
		ORDER BY path, created_at`

	// Get parent path first
	parent, err := r.GetCommentByID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent comment: %w", err)
	}

	pathPattern := parent.Path + ".%"
	maxAllowedDepth := parent.Depth + maxDepth

	comments := []*models.Comment{}
	err = r.getQueryable().SelectContext(ctx, &comments, query, pathPattern, maxAllowedDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment children: %w", err)
	}

	return comments, nil
}

// GetCommentTree builds a hierarchical tree structure
func (r *PostgresRepository) GetCommentTree(ctx context.Context, rootID string, maxDepth int, sortBy string) ([]*models.CommentTree, error) {
	// Get all comments for the root up to maxDepth
	filter := &models.CommentFilter{
		RootID:   &rootID,
		MaxDepth: &maxDepth,
		SortBy:   sortBy,
	}

	comments, err := r.GetComments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Build the tree structure
	return r.buildCommentTree(comments), nil
}

// buildCommentTree converts flat comments to tree structure
func (r *PostgresRepository) buildCommentTree(comments []*models.Comment) []*models.CommentTree {
	commentMap := make(map[string]*models.CommentTree)
	var roots []*models.CommentTree

	// Create all nodes
	for _, comment := range comments {
		node := &models.CommentTree{
			Comment:  comment,
			Children: []*models.CommentTree{},
		}
		commentMap[comment.ID] = node
	}

	// Build relationships
	for _, comment := range comments {
		node := commentMap[comment.ID]
		if comment.ParentID != nil {
			if parent, exists := commentMap[*comment.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	return roots
}

// GetCommentPath retrieves the path from root to a specific comment
func (r *PostgresRepository) GetCommentPath(ctx context.Context, commentID string) ([]*models.Comment, error) {
	comment, err := r.GetCommentByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// Parse path and get all comments in the path
	pathParts := strings.Split(comment.Path, ".")
	if len(pathParts) == 0 {
		return []*models.Comment{comment}, nil
	}

	query := `
		SELECT id, root_id, parent_id, user_id, content, media_url, link_url,
		       upvotes, downvotes, score, depth, path, is_deleted, created_at, updated_at
		FROM comments 
		WHERE id = ANY($1) AND NOT is_deleted
		ORDER BY depth`

	comments := []*models.Comment{}
	err = r.getQueryable().SelectContext(ctx, &comments, query, pq.Array(pathParts))
	if err != nil {
		return nil, fmt.Errorf("failed to get comment path: %w", err)
	}

	return comments, nil
}

// Vote operations
func (r *PostgresRepository) CreateVote(ctx context.Context, vote *models.Vote) error {
	if vote.ID == "" {
		vote.ID = uuid.New().String()
	}

	query := `
		INSERT INTO votes (id, comment_id, user_id, vote_type, created_at, updated_at)
		VALUES ($1::uuid, $2::uuid, $3::varchar, $4::smallint, $5::timestamptz, $6::timestamptz)
		ON CONFLICT (comment_id, user_id) 
		DO UPDATE SET vote_type = $4::smallint, updated_at = $6::timestamptz`

	vote.CreatedAt = time.Now()
	vote.UpdatedAt = time.Now()

	_, err := r.getDB().ExecContext(ctx, query,
		vote.ID, vote.CommentID, vote.UserID, vote.VoteType,
		vote.CreatedAt, vote.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}

	return nil
}

// UpdateVote updates or creates a vote
func (r *PostgresRepository) UpdateVote(ctx context.Context, commentID, userID string, voteType models.VoteType) error {
	vote := &models.Vote{
		CommentID: commentID,
		UserID:    userID,
		VoteType:  voteType,
	}
	return r.CreateVote(ctx, vote)
}

// DeleteVote removes a user's vote
func (r *PostgresRepository) DeleteVote(ctx context.Context, commentID, userID string) error {
	query := `DELETE FROM votes WHERE comment_id = $1 AND user_id = $2`

	_, err := r.getDB().ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}

	return nil
}

// GetUserVote retrieves a user's vote for a comment
func (r *PostgresRepository) GetUserVote(ctx context.Context, commentID, userID string) (*models.Vote, error) {
	query := `
		SELECT id, comment_id, user_id, vote_type, created_at, updated_at
		FROM votes 
		WHERE comment_id = $1 AND user_id = $2`

	vote := &models.Vote{}
	err := r.getQueryable().GetContext(ctx, vote, query, commentID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vote found
		}
		return nil, fmt.Errorf("failed to get user vote: %w", err)
	}

	return vote, nil
}

// GetCommentVotes retrieves all votes for a comment
func (r *PostgresRepository) GetCommentVotes(ctx context.Context, commentID string) ([]*models.Vote, error) {
	query := `
		SELECT id, comment_id, user_id, vote_type, created_at, updated_at
		FROM votes 
		WHERE comment_id = $1`

	votes := []*models.Vote{}
	err := r.getQueryable().SelectContext(ctx, &votes, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment votes: %w", err)
	}

	return votes, nil
}

// GetCommentsWithUserVotes retrieves comments with user's votes in a single query
func (r *PostgresRepository) GetCommentsWithUserVotes(ctx context.Context, rootID, userID string, filter *models.CommentFilter) ([]*models.Comment, map[string]*models.Vote, error) {
	query := `
		SELECT c.id, c.root_id, c.parent_id, c.user_id, c.content, c.media_url, c.link_url,
		       c.upvotes, c.downvotes, c.score, c.depth, c.path, c.is_deleted, c.created_at, c.updated_at,
		       v.id as vote_id, v.vote_type
		FROM comments c
		LEFT JOIN votes v ON c.id = v.comment_id AND v.user_id = $2
		WHERE c.root_id = $1 AND NOT c.is_deleted`

	args := []interface{}{rootID, userID}
	argIndex := 3

	if filter != nil {
		if filter.MaxDepth != nil {
			query += fmt.Sprintf(" AND c.depth <= $%d", argIndex)
			args = append(args, *filter.MaxDepth)
			argIndex++
		}

		// Add sorting
		sortBy := "c.created_at"
		if filter.SortBy != "" {
			switch filter.SortBy {
			case "score", "created_at", "updated_at":
				sortBy = "c." + filter.SortBy
			}
		}

		sortOrder := "DESC"
		if filter.SortOrder == "asc" {
			sortOrder = "ASC"
		}

		query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

		// Add pagination
		if filter.Limit != nil {
			query += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, *filter.Limit)
			argIndex++
		}

		if filter.Offset != nil {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, *filter.Offset)
			argIndex++
		}
	}

	rows, err := r.getDB().QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get comments with votes: %w", err)
	}
	defer rows.Close()

	comments := []*models.Comment{}
	votes := make(map[string]*models.Vote)

	for rows.Next() {
		comment := &models.Comment{}
		var voteID sql.NullString
		var voteType sql.NullInt32

		err := rows.Scan(
			&comment.ID, &comment.RootID, &comment.ParentID, &comment.UserID,
			&comment.Content, &comment.MediaURL, &comment.LinkURL,
			&comment.Upvotes, &comment.Downvotes, &comment.Score,
			&comment.Depth, &comment.Path, &comment.IsDeleted,
			&comment.CreatedAt, &comment.UpdatedAt,
			&voteID, &voteType,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan comment row: %w", err)
		}

		comments = append(comments, comment)

		if voteID.Valid {
			votes[comment.ID] = &models.Vote{
				ID:        voteID.String,
				CommentID: comment.ID,
				UserID:    userID,
				VoteType:  models.VoteType(voteType.Int32),
			}
		}
	}

	return comments, votes, nil
}

// UpdateCommentScores recalculates scores for specified comments
func (r *PostgresRepository) UpdateCommentScores(ctx context.Context, commentIDs []string) error {
	if len(commentIDs) == 0 {
		return nil
	}

	query := `
		UPDATE comments 
		SET 
			upvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = comments.id AND vote_type = 1),
			downvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = comments.id AND vote_type = -1),
			updated_at = NOW()
		WHERE id = ANY($1)`

	_, err := r.getDB().ExecContext(ctx, query, pq.Array(commentIDs))
	if err != nil {
		return fmt.Errorf("failed to update comment scores: %w", err)
	}

	// Update calculated score
	scoreQuery := `UPDATE comments SET score = upvotes - downvotes WHERE id = ANY($1)`
	_, err = r.getDB().ExecContext(ctx, scoreQuery, pq.Array(commentIDs))
	if err != nil {
		return fmt.Errorf("failed to update calculated scores: %w", err)
	}

	return nil
}

// GetCommentStats retrieves statistics for a root
func (r *PostgresRepository) GetCommentStats(ctx context.Context, rootID string) (*models.CommentStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_count,
			COALESCE(SUM(score), 0) as total_score,
			COALESCE(MAX(depth), 0) as max_depth,
			COUNT(CASE WHEN created_at > NOW() - INTERVAL '24 hours' THEN 1 END) as recent_count
		FROM comments 
		WHERE root_id = $1 AND NOT is_deleted`

	stats := &models.CommentStats{RootID: rootID}
	err := r.getQueryable().QueryRowxContext(ctx, query, rootID).Scan(
		&stats.TotalCount, &stats.TotalScore, &stats.MaxDepth, &stats.RecentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment stats: %w", err)
	}

	return stats, nil
}

// GetUserCommentCount retrieves the number of comments by a user
func (r *PostgresRepository) GetUserCommentCount(ctx context.Context, userID string) (int64, error) {
	query := `SELECT COUNT(*) FROM comments WHERE user_id = $1 AND NOT is_deleted`

	var count int64
	err := r.getQueryable().QueryRowxContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user comment count: %w", err)
	}

	return count, nil
}

// GetTopComments retrieves top comments based on score within time range
func (r *PostgresRepository) GetTopComments(ctx context.Context, rootID string, limit int, timeRange string) ([]*models.Comment, error) {
	var timeClause string
	switch timeRange {
	case "hour":
		timeClause = "AND created_at > NOW() - INTERVAL '1 hour'"
	case "day":
		timeClause = "AND created_at > NOW() - INTERVAL '1 day'"
	case "week":
		timeClause = "AND created_at > NOW() - INTERVAL '1 week'"
	case "month":
		timeClause = "AND created_at > NOW() - INTERVAL '1 month'"
	default:
		timeClause = "" // All time
	}

	query := fmt.Sprintf(`
		SELECT id, root_id, parent_id, user_id, content, media_url, link_url,
		       upvotes, downvotes, score, depth, path, is_deleted, created_at, updated_at
		FROM comments 
		WHERE root_id = $1 AND NOT is_deleted %s
		ORDER BY score DESC, created_at DESC
		LIMIT $2`, timeClause)

	comments := []*models.Comment{}
	err := r.getQueryable().SelectContext(ctx, &comments, query, rootID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top comments: %w", err)
	}

	return comments, nil
}

// PurgeDeletedComments permanently deletes soft-deleted comments older than specified days
func (r *PostgresRepository) PurgeDeletedComments(ctx context.Context, olderThan int) (int64, error) {
	query := `DELETE FROM comments WHERE is_deleted = true AND updated_at < NOW() - INTERVAL '%d days'`

	result, err := r.getDB().ExecContext(ctx, fmt.Sprintf(query, olderThan))
	if err != nil {
		return 0, fmt.Errorf("failed to purge deleted comments: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// RecalculateCommentScores recalculates all comment scores
func (r *PostgresRepository) RecalculateCommentScores(ctx context.Context) error {
	query := `
		UPDATE comments 
		SET 
			upvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = comments.id AND vote_type = 1),
			downvotes = (SELECT COUNT(*) FROM votes WHERE comment_id = comments.id AND vote_type = -1),
			updated_at = NOW()`

	_, err := r.getDB().ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to recalculate vote counts: %w", err)
	}

	// Update calculated scores
	scoreQuery := `UPDATE comments SET score = upvotes - downvotes`
	_, err = r.getDB().ExecContext(ctx, scoreQuery)
	if err != nil {
		return fmt.Errorf("failed to recalculate scores: %w", err)
	}

	return nil
}

// Transaction support
func (r *PostgresRepository) BeginTx(ctx context.Context) (repository.Repository, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresRepository{db: r.db, tx: tx}, nil
}

func (r *PostgresRepository) CommitTx(ctx context.Context) error {
	if r.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}

	err := r.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.tx = nil
	return nil
}

func (r *PostgresRepository) RollbackTx(ctx context.Context) error {
	if r.tx == nil {
		return fmt.Errorf("no transaction to rollback")
	}

	err := r.tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	r.tx = nil
	return nil
}
