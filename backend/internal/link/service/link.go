// Package service implements link use cases.
package service

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/repository"
)

// CreateLinkInput groups all parameters for creating a link.
type CreateLinkInput struct {
	Name        string
	Description *string
	Avatar      *string
	URL         string
}

// LinkService defines the interface for link business logic operations.
type LinkService interface {
	GetLinks(ctx context.Context) ([]*linkmodel.Link, error)
	CreateLink(ctx context.Context, input *CreateLinkInput) error
	CheckLinkStatus(ctx context.Context) error
	// GetOverview(ctx context.Context) (*response.LinkOverview, error)
}

// linkService implements the LinkService interface.
type linkService struct {
	linkRepo repository.LinkRepo
}

// NewLinkService creates a new link service instance.
func NewLinkService(linkRepo repository.LinkRepo) LinkService {
	return &linkService{linkRepo: linkRepo}
}

// GetLinks retrieves all enabled links.
func (s *linkService) GetLinks(ctx context.Context) ([]*linkmodel.Link, error) {
	return s.linkRepo.GetAllEnabled(ctx)
}

// CreateLink creates a new link.
func (s *linkService) CreateLink(ctx context.Context, input *CreateLinkInput) error {
	link := &linkmodel.Link{
		Name:        input.Name,
		Description: input.Description,
		Avatar:      input.Avatar,
		URL:         input.URL,
	}
	_, err := s.linkRepo.Create(ctx, link)
	return err
}

// CheckLinkStatus checks the status of all links by making HTTP requests.
// Updates the link status in the database if it has changed.
func (s linkService) CheckLinkStatus(ctx context.Context) error {
	links, err := s.linkRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	updates := make(map[uint]linkmodel.LinkStatus)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, 10) // Limit to 10 concurrent requests

	for _, link := range links {
		wg.Add(1)
		sem <- struct{}{}

		go func(l *linkmodel.Link) {
			defer wg.Done()
			defer func() { <-sem }()

			currentStatus := l.Status
			status := linkmodel.LinkStatusAbnormal

			// Check HTTPS links only
			if strings.HasPrefix(l.URL, "https://") {
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(l.URL)
				if err == nil && resp.StatusCode == http.StatusOK {
					status = linkmodel.LinkStatusNormal
				}
				if resp != nil {
					_ = resp.Body.Close()
				}
			}

			// Update status if changed
			if status != currentStatus {
				mu.Lock()
				updates[l.ID] = status
				mu.Unlock()
			}
		}(link)
	}

	wg.Wait()
	err = s.linkRepo.UpdateStatusBatch(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

// // GetOverview retrieves link statistics
// func (s *linkService) GetOverview(ctx context.Context) (*response.LinkOverview, error) {
// 	return s.linkRepo.GetOverview(ctx, s.db)
// }
