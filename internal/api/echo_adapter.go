package api

import (
	"context"
	"net/http"

	"github.com/christopher18/commentific/internal/service"
	"github.com/labstack/echo/v4"
)

// EchoAdapter wraps the CommentHandler for Echo framework
type EchoAdapter struct {
	handler *CommentHandler
}

// NewEchoAdapter creates a new Echo adapter for Commentific
func NewEchoAdapter(commentService *service.CommentService) *EchoAdapter {
	return &EchoAdapter{
		handler: NewCommentHandler(commentService),
	}
}

// RegisterRoutes registers all Commentific routes with an Echo instance
func (a *EchoAdapter) RegisterRoutes(e *echo.Echo) {
	// Create API group
	api := e.Group("/api/v1")

	// Comment operations
	api.POST("/comments", a.CreateComment)
	api.GET("/comments/:id", a.GetComment)
	api.PUT("/comments/:id", a.UpdateComment)
	api.DELETE("/comments/:id", a.DeleteComment)
	api.GET("/comments/:id/path", a.GetCommentPath)
	api.GET("/comments/:id/children", a.GetCommentChildren)

	// Voting operations
	api.POST("/comments/:id/vote", a.VoteComment)
	api.DELETE("/comments/:id/vote", a.RemoveVote)

	// Root-based operations
	api.GET("/roots/:root_id/comments", a.GetCommentsByRoot)
	api.GET("/roots/:root_id/comments/with-votes", a.GetCommentsWithVotes)
	api.GET("/roots/:root_id/tree", a.GetCommentTree)
	api.GET("/roots/:root_id/stats", a.GetCommentStats)
	api.GET("/roots/:root_id/top", a.GetTopComments)
	api.GET("/roots/:root_id/search", a.SearchComments)

	// User operations
	api.GET("/users/:user_id/comments", a.GetCommentsByUser)
	api.GET("/users/:user_id/count", a.GetUserCommentCount)

	// Health check
	e.GET("/health", a.HealthCheck)
}

// RegisterRoutesWithPrefix registers routes with a custom prefix
func (a *EchoAdapter) RegisterRoutesWithPrefix(e *echo.Echo, prefix string) {
	// Create API group with custom prefix
	api := e.Group(prefix)

	// Comment operations
	api.POST("/comments", a.CreateComment)
	api.GET("/comments/:id", a.GetComment)
	api.PUT("/comments/:id", a.UpdateComment)
	api.DELETE("/comments/:id", a.DeleteComment)
	api.GET("/comments/:id/path", a.GetCommentPath)
	api.GET("/comments/:id/children", a.GetCommentChildren)

	// Voting operations
	api.POST("/comments/:id/vote", a.VoteComment)
	api.DELETE("/comments/:id/vote", a.RemoveVote)

	// Root-based operations
	api.GET("/roots/:root_id/comments", a.GetCommentsByRoot)
	api.GET("/roots/:root_id/comments/with-votes", a.GetCommentsWithVotes)
	api.GET("/roots/:root_id/tree", a.GetCommentTree)
	api.GET("/roots/:root_id/stats", a.GetCommentStats)
	api.GET("/roots/:root_id/top", a.GetTopComments)
	api.GET("/roots/:root_id/search", a.SearchComments)

	// User operations
	api.GET("/users/:user_id/comments", a.GetCommentsByUser)
	api.GET("/users/:user_id/count", a.GetUserCommentCount)
}

// Echo handler adapters - these convert Echo contexts to http.Request/ResponseWriter

func (a *EchoAdapter) CreateComment(c echo.Context) error {
	a.handler.CreateComment(c.Response().Writer, c.Request())
	return nil
}

func (a *EchoAdapter) GetComment(c echo.Context) error {
	// Convert Echo params to mux.Vars format
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.GetComment(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) UpdateComment(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.UpdateComment(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) DeleteComment(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.DeleteComment(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentPath(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.GetCommentPath(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentChildren(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.GetCommentChildren(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) VoteComment(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.VoteComment(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) RemoveVote(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"id": c.Param("id")})
	a.handler.RemoveVote(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentsByRoot(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.GetCommentsByRoot(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentsWithVotes(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.GetCommentsWithVotes(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentTree(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.GetCommentTree(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentStats(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.GetCommentStats(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetTopComments(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.GetTopComments(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) SearchComments(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"root_id": c.Param("root_id")})
	a.handler.SearchComments(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetCommentsByUser(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"user_id": c.Param("user_id")})
	a.handler.GetCommentsByUser(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) GetUserCommentCount(c echo.Context) error {
	req := c.Request()
	req = addMuxVars(req, map[string]string{"user_id": c.Param("user_id")})
	a.handler.GetUserCommentCount(c.Response().Writer, req)
	return nil
}

func (a *EchoAdapter) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "commentific",
		"version":   "1.0.1",
		"timestamp": "2024-01-01T00:00:00Z", // You might want to use time.Now()
	})
}

// Helper function to add mux-style vars to request context
// This allows the existing handlers to work with Echo parameters
func addMuxVars(req *http.Request, vars map[string]string) *http.Request {
	// This is a simplified approach - in production you might want to use
	// a more robust method to pass parameters
	ctx := req.Context()
	for key, value := range vars {
		ctx = context.WithValue(ctx, key, value)
	}
	return req.WithContext(ctx)
}
