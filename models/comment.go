package models

import (
	"time"
)

// Comment represents a comment in the system with support for infinite hierarchy
type Comment struct {
	ID               string     `json:"id" db:"id"`
	RootID           string     `json:"root_id" db:"root_id"`                             // The entity this comment belongs to (post, product, etc.)
	ParentID         *string    `json:"parent_id" db:"parent_id"`                         // Parent comment ID for threading
	UserID           string     `json:"user_id" db:"user_id"`                             // External user ID
	Content          string     `json:"content" db:"content"`                             // The comment text
	MediaURL         *string    `json:"media_url" db:"media_url"`                         // Optional media attachment
	LinkURL          *string    `json:"link_url" db:"link_url"`                           // Optional link
	Upvotes          int64      `json:"upvotes" db:"upvotes"`                             // Number of upvotes
	Downvotes        int64      `json:"downvotes" db:"downvotes"`                         // Number of downvotes
	Score            int64      `json:"score" db:"score"`                                 // Calculated score (upvotes - downvotes)
	Depth            int        `json:"depth" db:"depth"`                                 // Depth in the comment tree
	Path             string     `json:"path" db:"path"`                                   // Materialized path for efficient queries
	IsDeleted        bool       `json:"is_deleted" db:"is_deleted"`                       // Soft delete flag
	IsEdited         bool       `json:"is_edited" db:"is_edited"`                         // Whether comment has been edited
	EditCount        int        `json:"edit_count" db:"edit_count"`                       // Number of times comment has been edited
	OriginalContent  *string    `json:"original_content,omitempty" db:"original_content"` // Original content before first edit
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	ContentUpdatedAt *time.Time `json:"content_updated_at,omitempty" db:"content_updated_at"` // When content was last edited
}

// Vote represents a user's vote on a comment
type Vote struct {
	ID        string    `json:"id" db:"id"`
	CommentID string    `json:"comment_id" db:"comment_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	VoteType  VoteType  `json:"vote_type" db:"vote_type"` // 1 for upvote, -1 for downvote
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// VoteType represents the type of vote
type VoteType int

const (
	VoteTypeNone VoteType = 0
	VoteTypeUp   VoteType = 1
	VoteTypeDown VoteType = -1
)

// CommentTree represents a comment with its children for hierarchical display
type CommentTree struct {
	Comment  *Comment       `json:"comment"`
	Children []*CommentTree `json:"children,omitempty"`
}

// CreateCommentRequest represents the request to create a new comment
type CreateCommentRequest struct {
	RootID   string  `json:"root_id" validate:"required"`
	ParentID *string `json:"parent_id"`
	UserID   string  `json:"user_id" validate:"required"`
	Content  string  `json:"content" validate:"required,min=1,max=10000"`
	MediaURL *string `json:"media_url"`
	LinkURL  *string `json:"link_url"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Content  *string `json:"content,omitempty"`
	MediaURL *string `json:"media_url,omitempty"`
	LinkURL  *string `json:"link_url,omitempty"`
}

// VoteRequest represents a vote request
type VoteRequest struct {
	UserID   string   `json:"user_id" validate:"required"`
	VoteType VoteType `json:"vote_type" validate:"required,oneof=1 -1"`
}

// CommentFilter represents filters for querying comments
type CommentFilter struct {
	RootID    *string `json:"root_id,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
	MaxDepth  *int    `json:"max_depth,omitempty"`
	SortBy    string  `json:"sort_by,omitempty"`    // "score", "created_at", "updated_at", "content_updated_at", "edit_count"
	SortOrder string  `json:"sort_order,omitempty"` // "asc", "desc"
	Limit     *int    `json:"limit,omitempty"`
	Offset    *int    `json:"offset,omitempty"`
	IsEdited  *bool   `json:"is_edited,omitempty"` // Filter by edited status
	MinEdits  *int    `json:"min_edits,omitempty"` // Minimum number of edits
	MaxEdits  *int    `json:"max_edits,omitempty"` // Maximum number of edits
}

// CommentStats represents statistics for a comment thread
type CommentStats struct {
	RootID             string  `json:"root_id"`
	TotalCount         int64   `json:"total_count"`
	TotalScore         int64   `json:"total_score"`
	MaxDepth           int     `json:"max_depth"`
	RecentCount        int64   `json:"recent_count"`          // Comments in last 24 hours
	EditedCount        int64   `json:"edited_count"`          // Number of edited comments
	TotalEdits         int64   `json:"total_edits"`           // Total number of edits across all comments
	EditRate           float64 `json:"edit_rate"`             // Percentage of comments that have been edited
	AvgEditsPerComment float64 `json:"avg_edits_per_comment"` // Average edits per edited comment
}
