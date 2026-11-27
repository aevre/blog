// Package service implements link use cases.
package service

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/authz"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/link/repository"
	"blog-server/internal/platform/field"
	"blog-server/internal/shared/contextx"
	"blog-server/internal/shared/errx"
)

// CreateInput groups all parameters for creating a link.
type CreateInput struct {
	Name        string
	Description *string
	Avatar      *string
	URL         string
	CategoryID  *uint
}

type UpdateInput struct {
	ID          uint
	Version     int64
	Name        field.Optional[string]
	URL         field.Optional[string]
	Description field.Optional[string]
	Avatar      field.Optional[string]
	SortOrder   field.Optional[int]
	CategoryID  field.Optional[uint]
}

type ReviewInput struct {
	ID       uint
	Version  int64
	Decision linkmodel.LinkReviewStatus
	Reason   *string
}

// LinkService defines the interface for link business logic operations.
type LinkService interface {
	ListPublic(ctx context.Context) ([]*linkmodel.Link, error)
	AdminList(ctx context.Context, actor contextx.User) ([]*linkmodel.Link, error)

	GetMineByID(ctx context.Context, actor contextx.User, id uint) (*linkmodel.Link, error)
	AdminGetByID(ctx context.Context, actor contextx.User, id uint) (*linkmodel.Link, error)

	Create(ctx context.Context, actor contextx.User, in CreateInput) (*linkmodel.Link, error)
	Update(ctx context.Context, actor contextx.User, in UpdateInput) (*linkmodel.Link, error)
	Delete(ctx context.Context, actor contextx.User, id uint, version int64) error
	Review(ctx context.Context, actor contextx.User, in ReviewInput) error

	CheckLinkStatus(ctx context.Context) error
}

// linkService implements the LinkService interface.
type linkService struct {
	linkRepo   repository.LinkRepo
	authorizer authz.Authorizer
}

// NewLinkService creates a new link service instance.
func NewLinkService(
	linkRepo repository.LinkRepo,
	authorizer authz.Authorizer,
) LinkService {
	return &linkService{
		linkRepo:   linkRepo,
		authorizer: authorizer,
	}
}

func (s *linkService) requestActor(
	actor contextx.User,
	resource authz.Resource,
	action authz.Action,
) error {
	if actor.ID == 0 {
		return errx.New(errx.CodeUnauthorized, errors.New("missing authenticated user"))
	}
	return s.authorizer.Require(actor.Role, resource, action)
}

func requireAdmin(user contextx.User) error {
	if user.Role != accountmodel.UserRoleAdmin {
		return authz.ErrForbidden()
	}
	return nil
}

// GetLinks retrieves all enabled links.
func (s *linkService) ListPublic(ctx context.Context) ([]*linkmodel.Link, error) {
	return s.linkRepo.ListPublic(ctx)
}

func (s *linkService) AdminList(ctx context.Context, actor contextx.User) ([]*linkmodel.Link, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionRead); err != nil {
		return nil, err
	}
	return s.linkRepo.ListAll(ctx)
}

func (s *linkService) GetMineByID(ctx context.Context, actor contextx.User, id uint) (*linkmodel.Link, error) {
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionRead); err != nil {
		return nil, err
	}

	return s.linkRepo.GetMineByID(ctx, actor.ID, id)
}

func (s *linkService) AdminGetByID(ctx context.Context, actor contextx.User, id uint) (*linkmodel.Link, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionRead); err != nil {
		return nil, err
	}
	return s.linkRepo.GetByID(ctx, id)
}

// Create creates a new link.
func (s *linkService) Create(ctx context.Context, actor contextx.User, in CreateInput) (*linkmodel.Link, error) {
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionCreate); err != nil {
		return nil, err
	}

	in.Name = strings.TrimSpace(in.Name)
	in.URL = strings.TrimSpace(in.URL)
	if in.Name == "" {
		return nil, errx.New(errx.CodeValidationFailed, errors.New("name is required"))
	}

	url, err := validateLinkURL(in.URL)
	if err != nil {
		return nil, errx.New(errx.CodeValidationFailed, err)
	}

	return s.linkRepo.Create(ctx,
		&repository.CreateParams{
			SubmitterID: actor.ID,
			Name:        in.Name,
			Description: in.Description,
			Avatar:      in.Avatar,
			URL:         url.String(),
		},
	)
}

func (s *linkService) Update(ctx context.Context, actor contextx.User, in UpdateInput) (*linkmodel.Link, error) {
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionUpdate); err != nil {
		return nil, err
	}
	if in.ID == 0 || in.Version < 1 {
		return nil, errx.New(errx.CodeInvalidParam, errors.New("invalid ID/version"))
	}
	if in.Name.Set && (in.Name.Value == nil || strings.TrimSpace(*in.Name.Value) == "") {
		return nil, errx.New(errx.CodeValidationFailed, errors.New("name cannot be empty"))
	}
	if in.URL.Set {
		if in.URL.Value == nil {
			return nil, errx.New(errx.CodeValidationFailed, errors.New("URL cannot be null"))
		}
		if _, err := validateLinkURL(strings.TrimSpace(*in.URL.Value)); err != nil {
			return nil, errx.New(errx.CodeValidationFailed, err)
		}
	}

	return s.linkRepo.UpdateContent(ctx, &repository.UpdateContentParams{
		ID:              in.ID,
		ActorID:         actor.ID,
		IsAdmin:         actor.Role == accountmodel.UserRoleAdmin,
		ExpectedVersion: in.Version,
		Name:            in.Name,
		URL:             in.URL,
		Description:     in.Description,
		Avatar:          in.Avatar,
		SortOrder:       in.SortOrder,
		CategoryID:      in.CategoryID,
	})
}

func (s *linkService) Delete(ctx context.Context, actor contextx.User, id uint, version int64) error {
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionDelete); err != nil {
		return err
	}
	if id == 0 || version < 1 {
		return errx.New(errx.CodeInvalidParam, errors.New("invalid link ID or version"))
	}
	if actor.Role == accountmodel.UserRoleAdmin {
		return s.linkRepo.DeleteAdmin(ctx, id, version)
	}
	return s.linkRepo.DeleteMine(ctx, actor.ID, id, version)
}

func (s *linkService) Review(ctx context.Context, actor contextx.User, in ReviewInput) error {
	if err := s.requestActor(actor, authz.ResourceLink, authz.ActionReview); err != nil {
		return err
	}
	if err := requireAdmin(actor); err != nil {
		return err
	}

	if in.Decision != linkmodel.LinkReviewApproved && in.Decision != linkmodel.LinkReviewRejected {
		return errx.New(errx.CodeInvalidParam, errors.New("invalid decision"))
	}

	if in.Decision == linkmodel.LinkReviewRejected && (in.Reason == nil || strings.TrimSpace(*in.Reason) == "") {
		return errx.New(errx.CodeInvalidParam, errors.New("rejection reason required"))
	}

	return s.linkRepo.Review(ctx, &repository.ReviewParams{
		ID:              in.ID,
		ExpectedVersion: in.Version,
		Decision:        in.Decision,
		Reason:          in.Reason,
	})
}

// CheckLinkStatus checks the status of all links by making HTTP requests.
// Updates the link status in the database if it has changed.
func (s linkService) CheckLinkStatus(ctx context.Context) error {
	links, err := s.linkRepo.ListAll(ctx)
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
	err = s.linkRepo.UpdateHealthStatusBatch(ctx, updates)
	if err != nil {
		return err
	}

	return nil
}

func validateLinkURL(raw string) (*url.URL, error) {
	if raw == "" || len(raw) > 255 {
		return nil, errors.New("invalid link URL length")
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Opaque != "" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("link URL must be an absolute HTTPS URL")
	}

	if port := u.Port(); port != "" && port != "443" {
		return nil, errors.New("link URL must use HTTPS port 443")
	}

	if _, err := netip.ParseAddr(u.Hostname()); err == nil {
		return nil, errors.New("ID address URLs are not allowed")
	}

	return u, nil
}
