// Package service implements summary generation sessions.
package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"blog-server/internal/platform/config"
	"blog-server/internal/platform/logging"
	"blog-server/internal/shared/errx"
)

// ModelService defines the interface for LLM-related operations.
type ModelService interface {
	GenerateSummary(ctx context.Context, content string) (<-chan string, <-chan error)
}

// modelService implements the ModelService interface.
type modelService struct {
	cfg *config.Config
	log logger.Logger
}

// NewModelService creates and returns a new ModelService instance.
func NewModelService(cfg *config.Config, log logger.Logger) ModelService {
	return &modelService{cfg: cfg, log: log}
}

// GenerateSummary generates a summary for the given content using LLM streaming.
func (s *modelService) GenerateSummary(ctx context.Context, content string) (<-chan string, <-chan error) {
	textCh := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(textCh)
		defer close(errCh)

		payload := map[string]any{
			"model": "gpt-4.1-mini",
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": "请为我总结这篇文章（以一百字以内的纯文本格式返回）：" + content,
				},
			},
			"max_tokens": 8000,
			"stream":     true,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			errCh <- errx.New(errx.CodeInternalError, fmt.Errorf("marshal llm request payload: %w", err))
			return
		}
		req, err := http.NewRequestWithContext(ctx, "POST",
			"https://models.inference.ai.azure.com/chat/completions", bytes.NewBuffer(body))
		if err != nil {
			errCh <- errx.New(errx.CodeExternalError, fmt.Errorf("llm request error: %w", err))
			return
		}
		req.Header.Set("Authorization", "Bearer "+s.cfg.LLM.APIKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 0}
		resp, err := client.Do(req)
		if err != nil {
			errCh <- errx.New(errx.CodeExternalError, fmt.Errorf("llm request error: %w", err))
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err = errx.New(errx.CodeExternalError, fmt.Errorf("status code: %d", resp.StatusCode))
			errCh <- err
			s.log.Error("llm request failed",
				logger.String("module", "GenerateSummary"),
				logger.Err(err),
				logger.Int("status_code", resp.StatusCode),
			)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 1024*1024)

		doneRegexp := regexp.MustCompile(`data:\s*\[DONE\]`)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				s.log.Info("[GenerateSummary] context canceled")
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			if doneRegexp.MatchString(line) {
				s.log.Info("[GenerateSummary] stream finished")
				textCh <- "[DONE]"
				return
			}

			jsonStr := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if jsonStr == "" {
				continue
			}

			var chunk map[string]any
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				s.log.Warn("invalid llm stream chunk",
					logger.String("module", "GenerateSummary"),
					logger.Err(err),
				)
				continue
			}

			choices, ok := chunk["choices"].([]any)
			if !ok || len(choices) == 0 {
				continue
			}
			choice, ok := choices[0].(map[string]any)
			if !ok {
				continue
			}
			delta, ok := choice["delta"].(map[string]any)
			if !ok {
				continue
			}
			text, ok := delta["content"].(string)
			if !ok || text == "" {
				continue
			}

			select {
			case <-ctx.Done():
				s.log.Info("[GenerateSummary] context canceled during sending")
				return
			case textCh <- text:
			}
		}

		if err := scanner.Err(); err != nil {
			s.log.Error("scanner error",
				logger.String("module", "GenerateSummary"),
				logger.Err(err),
			)
			errCh <- err
		}
	}()

	return textCh, errCh
}
