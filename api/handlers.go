package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/christopher18/commentific/v2/models"
	"github.com/christopher18/commentific/v2/service"
	"github.com/gorilla/mux"
)

// CommentHandler handles HTTP requests for comment operations
type CommentHandler struct {
	commentService *service.CommentService
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total,omitempty"`
}

// Helper functions

func (h *CommentHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *CommentHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	h.sendJSONResponse(w, statusCode, response)
}

func (h *CommentHandler) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	h.sendJSONResponse(w, http.StatusOK, response)
}

// getUserID extracts user ID from request headers or query params
func (h *CommentHandler) getUserID(r *http.Request) string {
	// In production, this would typically come from JWT token or session
	// For now, we'll use a header or query parameter
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("user_id")
	}
	return userID
}

// parseCommentFilter parses query parameters into CommentFilter
func (h *CommentHandler) parseCommentFilter(r *http.Request) *models.CommentFilter {
	filter := &models.CommentFilter{}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = &l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filter.Offset = &o
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	if maxDepth := r.URL.Query().Get("max_depth"); maxDepth != "" {
		if d, err := strconv.Atoi(maxDepth); err == nil {
			filter.MaxDepth = &d
		}
	}

	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		filter.ParentID = &parentID
	}

	return filter
}

// CreateComment handles POST /comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCommentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// If user_id not in request body, try to get from headers/query
	if req.UserID == "" {
		req.UserID = h.getUserID(r)
		if req.UserID == "" {
			h.sendErrorResponse(w, http.StatusBadRequest, "User ID is required")
			return
		}
	}

	comment, err := h.commentService.CreateComment(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "validation failed") {
			h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendJSONResponse(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    comment,
		Message: "Comment created successfully",
	})
}

// GetComment handles GET /comments/{id}
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	comment, err := h.commentService.GetComment(r.Context(), commentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, "Comment not found")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, comment)
}

// UpdateComment handles PUT /comments/{id}
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	userID := h.getUserID(r)

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	if userID == "" {
		h.sendErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	var req models.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	err := h.commentService.UpdateComment(r.Context(), commentID, userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			h.sendErrorResponse(w, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Comment updated successfully",
	})
}

// DeleteComment handles DELETE /comments/{id}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	userID := h.getUserID(r)

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	if userID == "" {
		h.sendErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	err := h.commentService.DeleteComment(r.Context(), commentID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			h.sendErrorResponse(w, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Comment deleted successfully",
	})
}

// GetCommentsByRoot handles GET /roots/{root_id}/comments
func (h *CommentHandler) GetCommentsByRoot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	filter := h.parseCommentFilter(r)
	comments, err := h.commentService.GetCommentsByRoot(r.Context(), rootID, filter)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := PaginatedResponse{
		Success: true,
		Data:    comments,
	}

	if filter.Limit != nil && filter.Offset != nil {
		response.Pagination = &Pagination{
			Limit:  *filter.Limit,
			Offset: *filter.Offset,
		}
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetCommentTree handles GET /roots/{root_id}/tree
func (h *CommentHandler) GetCommentTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	maxDepth := 10 // default
	if d := r.URL.Query().Get("max_depth"); d != "" {
		if depth, err := strconv.Atoi(d); err == nil {
			maxDepth = depth
		}
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "score"
	}

	tree, err := h.commentService.GetCommentTree(r.Context(), rootID, maxDepth, sortBy)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, tree)
}

// GetCommentsByUser handles GET /users/{user_id}/comments
func (h *CommentHandler) GetCommentsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	filter := h.parseCommentFilter(r)
	comments, err := h.commentService.GetCommentsByUser(r.Context(), userID, filter)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := PaginatedResponse{
		Success: true,
		Data:    comments,
	}

	if filter.Limit != nil && filter.Offset != nil {
		response.Pagination = &Pagination{
			Limit:  *filter.Limit,
			Offset: *filter.Offset,
		}
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// VoteComment handles POST /comments/{id}/vote
func (h *CommentHandler) VoteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	userID := h.getUserID(r)

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	if userID == "" {
		h.sendErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	var req models.VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Override user ID from auth context
	req.UserID = userID

	err := h.commentService.VoteComment(r.Context(), commentID, userID, req.VoteType)
	if err != nil {
		if strings.Contains(err.Error(), "cannot vote on their own") {
			h.sendErrorResponse(w, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Vote recorded successfully",
	})
}

// RemoveVote handles DELETE /comments/{id}/vote
func (h *CommentHandler) RemoveVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]
	userID := h.getUserID(r)

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	if userID == "" {
		h.sendErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	err := h.commentService.RemoveVote(r.Context(), commentID, userID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Vote removed successfully",
	})
}

// GetCommentsWithVotes handles GET /roots/{root_id}/comments/with-votes
func (h *CommentHandler) GetCommentsWithVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]
	userID := h.getUserID(r)

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	if userID == "" {
		h.sendErrorResponse(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	filter := h.parseCommentFilter(r)
	comments, votes, err := h.commentService.GetCommentsWithUserVotes(r.Context(), rootID, userID, filter)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	data := map[string]interface{}{
		"comments": comments,
		"votes":    votes,
	}

	response := PaginatedResponse{
		Success: true,
		Data:    data,
	}

	if filter.Limit != nil && filter.Offset != nil {
		response.Pagination = &Pagination{
			Limit:  *filter.Limit,
			Offset: *filter.Offset,
		}
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// GetCommentStats handles GET /roots/{root_id}/stats
func (h *CommentHandler) GetCommentStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	stats, err := h.commentService.GetCommentStats(r.Context(), rootID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, stats)
}

// GetTopComments handles GET /roots/{root_id}/top
func (h *CommentHandler) GetTopComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "day"
	}

	comments, err := h.commentService.GetTopComments(r.Context(), rootID, limit, timeRange)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, comments)
}

// SearchComments handles GET /roots/{root_id}/search
func (h *CommentHandler) SearchComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rootID := vars["root_id"]
	query := r.URL.Query().Get("q")

	if rootID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Root ID is required")
		return
	}

	if query == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Search query is required")
		return
	}

	filter := h.parseCommentFilter(r)
	comments, err := h.commentService.SearchComments(r.Context(), rootID, query, filter)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, comments)
}

// GetUserCommentCount handles GET /users/{user_id}/count
func (h *CommentHandler) GetUserCommentCount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	count, err := h.commentService.GetUserCommentCount(r.Context(), userID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, map[string]interface{}{
		"user_id": userID,
		"count":   count,
	})
}

// GetCommentPath handles GET /comments/{id}/path
func (h *CommentHandler) GetCommentPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	path, err := h.commentService.GetCommentPath(r.Context(), commentID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, "Comment not found")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, path)
}

// GetCommentChildren handles GET /comments/{id}/children
func (h *CommentHandler) GetCommentChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["id"]

	if commentID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	// Parse max_depth parameter
	maxDepth := 10 // default
	if maxDepthStr := r.URL.Query().Get("max_depth"); maxDepthStr != "" {
		if parsed, err := strconv.Atoi(maxDepthStr); err == nil && parsed > 0 {
			maxDepth = parsed
		}
	}

	children, err := h.commentService.GetCommentChildren(r.Context(), commentID, maxDepth)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendErrorResponse(w, http.StatusNotFound, "Comment not found")
		} else {
			h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.sendSuccessResponse(w, children)
}
