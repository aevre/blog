// Package handler exposes summary HTTP and streaming endpoints.
package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"blog-server/internal/platform/logging"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/httpx"
	"blog-server/internal/summary/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// ModelHandler defines the interface for model HTTP handlers.
type ModelHandler interface {
	CreateSummarySession(c *echo.Context) error
	SummaryStream(c *echo.Context) error
}

// Session represents an SSE streaming session.
type Session struct {
	Content  string
	TaskType string
	TextCh   <-chan string
	ErrCh    <-chan error
}

// modelHandler implements the ModelHandler interface.
type modelHandler struct {
	log      logger.Logger
	svc      service.ModelService
	sessions *sync.Map
}

// NewModelHandler creates a new model handler instance.
func NewModelHandler(log logger.Logger, svc service.ModelService) ModelHandler {
	return &modelHandler{
		log:      log,
		svc:      svc,
		sessions: &sync.Map{},
	}
}

// CreateSummarySession creates a new summarization session.
func (h *modelHandler) CreateSummarySession(c *echo.Context) error {
	req := new(struct {
		Content string `json:"content" validate:"required"`
	})
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	sessionID := uuid.New().String()

	session := &Session{
		Content:  req.Content,
		TaskType: "summary",
		TextCh:   nil,
		ErrCh:    nil,
	}

	h.sessions.Store(sessionID, session)

	return httpx.OK(c, map[string]string{
		"sessionId": sessionID,
	})
}

// SummaryStream handles SSE streaming for summarization.
func (h *modelHandler) SummaryStream(c *echo.Context) error {
	sessionID := c.Param("sessionId")
	return h.stream(c, sessionID)
}

// stream handles the SSE streaming logic for a given session.
func (h *modelHandler) stream(c *echo.Context, sessionID string) error {
	v, ok := h.sessions.Load(sessionID)
	if !ok {
		h.log.Warn("session not found")
		return errx.New(errx.CodeNotFound, fmt.Errorf("session %s not found", sessionID))
	}
	session := v.(*Session)
	defer h.sessions.Delete(sessionID)

	if session.TextCh == nil && session.ErrCh == nil {
		switch session.TaskType {
		case "summary":
			session.TextCh, session.ErrCh = h.svc.GenerateSummary(c.Request().Context(), session.Content)
		default:
			return errx.New(errx.CodeInvalidParam, fmt.Errorf("unsupported task type"))
		}
		h.sessions.Store(sessionID, session)
	}

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	for {
		select {
		case <-c.Request().Context().Done():
			return nil

		case err := <-session.ErrCh:
			if err != nil {
				if _, werr := fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error()); werr != nil {
					return werr
				}
				return rc.Flush()
			}

		case text := <-session.TextCh:
			if text == "[DONE]" {
				if _, werr := fmt.Fprintf(w, "event: done\ndata: %s\n\n", text); werr != nil {
					return werr
				}
				return rc.Flush()
			}
			if text == "" {
				continue
			}
			for line := range strings.SplitSeq(text, "\n") {
				if _, werr := fmt.Fprintf(w, "data: %s\n\n", line); werr != nil {
					return werr
				}
			}
			if err := rc.Flush(); err != nil {
				return err
			}
		}
	}
}

// RegisterModelRoutes registers all model-related routes.
func RegisterModelRoutes(r *echo.Group, h ModelHandler) {
	group := r.Group("/model")
	group.POST("/summarize", h.CreateSummarySession)
	group.GET("/summarize/:sessionId", h.SummaryStream)
}
