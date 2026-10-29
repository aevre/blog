package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	contentmodel "blog-server/internal/content/model"
	"blog-server/internal/content/repository"
	"blog-server/internal/platform/config"
)

// RssService defines the interface for RSS feed generation operations.
type RssService interface {
	GenerateRSSFeed(ctx context.Context) (*contentmodel.RSS, error)
	GeneratePagedFeed(ctx context.Context, page, pageSize int) (*contentmodel.RSS, error)
	GenerateCompleteFeed(ctx context.Context) (*contentmodel.RSS, error)
}

// rssService implements the RssService interface.
type rssService struct {
	cfg      config.AppConfig
	postRepo repository.PostRepo
}

// NewRssService creates and returns a new RssService instance.
func NewRssService(cfg *config.Config, postRepo repository.PostRepo) RssService {
	return &rssService{cfg: cfg.App, postRepo: postRepo}
}

// GenerateRSSFeed generates the default RSS feed with the first page of posts.
func (s *rssService) GenerateRSSFeed(ctx context.Context) (*contentmodel.RSS, error) {
	defaultPageSize := 100
	return s.GeneratePagedFeed(ctx, 1, defaultPageSize)
}

// GeneratePagedFeed generates a paginated RSS feed.
func (s *rssService) GeneratePagedFeed(ctx context.Context, page, pageSize int) (*contentmodel.RSS, error) {
	total, _ := s.postRepo.CountPublished(ctx)

	if page < 1 {
		page = 1
	}

	posts, err := s.postRepo.ListPublishedForMeta(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	atomLinks := []contentmodel.AtomLink{
		{
			Href: fmt.Sprintf("%s/api/v1/rss?page=%d&pageSize=%d", s.cfg.Domain, page, pageSize),
			Rel:  "self",
			Type: "application/rss+xml",
		},
	}
	if hasNext {
		atomLinks = append(atomLinks, contentmodel.AtomLink{
			Href: fmt.Sprintf("%s/api/v1/rss?page=%d&pageSize=%d", s.cfg.Domain, page+1, pageSize),
			Rel:  "next",
			Type: "application/rss+xml",
		})
	}
	if hasPrev {
		atomLinks = append(atomLinks, contentmodel.AtomLink{
			Href: fmt.Sprintf("%s/api/v1/rss?page=%d&pageSize=%d", s.cfg.Domain, page-1, pageSize),
			Rel:  "previous",
			Type: "application/rss+xml",
		})
	}
	atomLinks = append(atomLinks,
		contentmodel.AtomLink{
			Href: fmt.Sprintf("%s/api/v1/rss?page=1&pageSize=%d", s.cfg.Domain, pageSize),
			Rel:  "first",
			Type: "application/rss+xml",
		},
		contentmodel.AtomLink{
			Href: fmt.Sprintf("%s/api/v1/rss?page=%d&pageSize=%d", s.cfg.Domain, totalPages, pageSize),
			Rel:  "last",
			Type: "application/rss+xml",
		},
	)

	feed := s.newBaseRSS(ctx)

	feed.Channel.AtomLinks = atomLinks
	feed.Channel.Description = fmt.Sprintf("Latest posts - Page %d of %d", page, totalPages)
	feed.Channel.Items = s.convertPostsToItems(posts)

	if page > 1 {
		feed.Channel.Complete = &contentmodel.FhComplete{}
	}

	return feed, nil
}

// GenerateCompleteFeed generates an RSS feed containing all published posts.
func (s *rssService) GenerateCompleteFeed(ctx context.Context) (*contentmodel.RSS, error) {
	total, err := s.postRepo.CountPublished(ctx)
	if err != nil || total == 0 {
		total = 500
	}

	posts, err := s.postRepo.ListPublishedForMeta(ctx, 1, total)
	if err != nil {
		return nil, err
	}

	feed := s.newBaseRSS(ctx)

	feed.Channel.Title = s.cfg.Name + " (Complete Archive)"
	feed.Channel.Description = "Full history archive of posts"
	feed.Channel.Items = s.convertPostsToItems(posts)

	feed.Channel.Complete = &contentmodel.FhComplete{}
	feed.Channel.AtomLinks = []contentmodel.AtomLink{
		{
			Href: s.cfg.Domain + "/api/v1/rss/complete",
			Rel:  "self",
			Type: "application/rss+xml",
		},
	}

	return feed, nil
}

// newBaseRSS creates a base RSS struct with common fields.
func (s *rssService) newBaseRSS(ctx context.Context) *contentmodel.RSS {
	return &contentmodel.RSS{
		Version: "2.0",
		XMLNs:   "http://www.w3.org/2005/Atom",
		Content: "http://purl.org/rss/1.0/modules/content/",
		DC:      "http://purl.org/dc/elements/1.1/",
		FH:      "http://purl.org/syndication/history/1.0",
		Channel: contentmodel.RssChannel{
			Title:     s.cfg.Name,
			Link:      s.cfg.Domain,
			LastBuild: s.getBuildTime(ctx),
		},
	}
}

// getBuildTime returns the latest published time as the build time.
func (s *rssService) getBuildTime(ctx context.Context) string {
	latestPublishedAt, err := s.postRepo.GetLatestPublishedAt(ctx)
	if err != nil || latestPublishedAt == nil {
		return time.Now().Format(time.RFC1123Z)
	}
	return latestPublishedAt.Format(time.RFC1123Z)
}

// convertPostsToItems converts posts to RSS items.
func (s *rssService) convertPostsToItems(posts []*contentmodel.Post) []contentmodel.RssItem {
	items := make([]contentmodel.RssItem, len(posts))
	for i, post := range posts {
		link := s.cfg.Domain + "/blog/" + strconv.Itoa(int(post.ID))

		summary := ""
		if post.Summary != nil {
			summary = *post.Summary
		}

		pubDate := post.PublishedAt
		if pubDate == nil {
			// CreatedAt is not selected by ListPublishedForMeta; UpdatedAt is.
			pubDate = &post.UpdatedAt
		}

		items[i] = contentmodel.RssItem{
			Title:       post.Title,
			Link:        link,
			GUID:        contentmodel.RssGUID{Value: link, IsPermaLink: true},
			PubDate:     pubDate.Format(time.RFC1123Z),
			Description: contentmodel.RssItemDescription{Value: summary},
		}
	}
	return items
}
