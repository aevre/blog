package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"blog-server/authz"
	"blog-server/cache"
	"blog-server/contextx"
	"blog-server/entity"
	"blog-server/logger"
	"blog-server/pkg/txmgr"
	"blog-server/repository"
)

// PostService defines the interface for post business logic operations.
type PostService interface {
	GetPosts(ctx context.Context, page, pageSize int) ([]*entity.Post, int, error)
	GetPostsWithContent(ctx context.Context) ([]*entity.Post, error)
	GetPostsMeta(ctx context.Context) []*entity.Post
	GetPostByID(ctx context.Context, id uint) (*entity.Post, error)
	CreatePost(ctx context.Context, user contextx.User, input *CreatePostInput) (*entity.Post, error)
	FlushViewCountToDB(ctx context.Context) error

	AdminGetPosts(ctx context.Context, user contextx.User, status *entity.PostStatus, keyword *string, page, pageSize int) ([]*entity.Post, int, error)
	AdminGetPostByID(ctx context.Context, user contextx.User, id uint) (*entity.Post, error)
	UpdatePost(ctx context.Context, user contextx.User, input *UpdatePostInput) (*entity.Post, error)
	DeletePost(ctx context.Context, user contextx.User, id uint) error
}

// CreatePostInput groups all parameters for creating a post.
type CreatePostInput struct {
	Title   string
	Summary *string
	Content string
	Cover   *string
	Status  entity.PostStatus

	UserID uint

	CategoryIDs []uint
	Tags        []uint
}

// UpdatePostInput groups all parameters for updating a post.
type UpdatePostInput struct {
	ID uint

	Title       *string
	Summary     *string
	Cover       *string
	Content     *string
	Status      *entity.PostStatus
	CategoryIDs *[]uint
	Tags        *[]uint
}

// postService implements the PostService interface.
type postService struct {
	tx    txmgr.TxManager
	log   logger.Logger
	rc    cache.CacheClient
	pr    repository.PostRepo
	authz *authz.Authorizer
}

// NewPostService creates and returns a new PostService instance.
func NewPostService(
	tx txmgr.TxManager,
	log logger.Logger,
	pr repository.PostRepo,
	rc cache.CacheClient,
	authz *authz.Authorizer,
) PostService {
	return &postService{
		log:   log,
		rc:    rc,
		tx:    tx,
		pr:    pr,
		authz: authz,
	}
}

// GetPosts retrieves all published posts with pagination.
func (s *postService) GetPosts(ctx context.Context, page, pageSize int) ([]*entity.Post, int, error) {
	count, err := s.pr.CountPublished(ctx)
	if err != nil {
		return nil, 0, err
	}
	ps, err := s.pr.ListPublished(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return ps, count, nil
}

// GetPostsWithContent retrieves all posts with content.
// WARNING: The current repo does not support it yet, so reuse it for now.
func (s *postService) GetPostsWithContent(ctx context.Context) ([]*entity.Post, error) {
	return nil, nil
}

// GetPostsMeta retrieves metadata for all published posts.
func (s *postService) GetPostsMeta(ctx context.Context) []*entity.Post {
	posts, err := s.pr.ListPublishedForSitemap(ctx)
	if err != nil {
		s.log.Error("get posts meta failed", logger.Err(err))
		return []*entity.Post{}
	}
	return posts
}

// GetPostByID retrieves a single post and asynchronously increments view count.
func (s *postService) GetPostByID(ctx context.Context, id uint) (*entity.Post, error) {
	post, err := s.pr.GetPublishedByID(ctx, id)
	if err != nil {
		return nil, err
	}

	go func(postID uint) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := s.rc.Incr(bgCtx, fmt.Sprintf("blog:post:view_count:%d", postID)); err != nil {
			s.log.Error("incr post view count failed", logger.Err(err))
		}
	}(post.ID)

	return post, nil
}

// CreatePost creates a new post with tags and categories.
func (s *postService) CreatePost(ctx context.Context, user contextx.User, input *CreatePostInput) (*entity.Post, error) {
	if err := s.authz.Authorize(ctx, user.ID, user.Role, authz.ResourcePost, authz.ActionCreate, nil); err != nil {
		return nil, err
	}
	contentLength := utf8.RuneCountInString(input.Content)

	readTime := uint(max(1, (contentLength+199)/200))

	post := &entity.Post{
		Title:           input.Title,
		Summary:         input.Summary,
		Cover:           input.Cover,
		Content:         input.Content,
		ReadTimeMinutes: readTime,
		UserID:          input.UserID,
		Status:          input.Status,
	}
	var err error

	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		post, err = s.pr.Create(ctx, post)
		if err != nil {
			return err
		}

		if err = s.pr.AddTags(ctx, post.ID, input.Tags); err != nil {
			return err
		}
		if err = s.pr.AddCategories(ctx, post.ID, input.CategoryIDs); err != nil {
			return err
		}

		post, err = s.pr.GetAdminListItemByID(ctx, post.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return post, nil
}

// AdminGetPosts retrieves all posts inclueing drafts for admin.
func (s *postService) AdminGetPosts(ctx context.Context, user contextx.User, status *entity.PostStatus, keyword *string, page, pageSize int) ([]*entity.Post, int, error) {
	if user.Role != entity.UserRoleAdmin {
		return nil, 0, authz.ErrForbidden()
	}
	count, err := s.pr.CountAll(ctx, status, keyword)
	if err != nil {
		return nil, 0, err
	}
	ps, err := s.pr.ListAll(ctx, status, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return ps, count, nil
}

// AdminGetPostByID retrieves a single post by ID for admin.
func (s *postService) AdminGetPostByID(ctx context.Context, user contextx.User, id uint) (*entity.Post, error) {
	if user.Role != entity.UserRoleAdmin {
		return nil, authz.ErrForbidden()
	}
	return s.pr.GetByID(ctx, id)
}

// UpdatePost updates a post with tags and categories.
func (s *postService) UpdatePost(ctx context.Context, user contextx.User, input *UpdatePostInput) (*entity.Post, error) {
	if err := s.authz.Authorize(ctx, user.ID, user.Role, authz.ResourcePost, authz.ActionUpdate, &input.ID); err != nil {
		return nil, err
	}

	post := &entity.Post{
		ID: input.ID,
	}

	if input.Title != nil {
		post.Title = *input.Title
	}
	if input.Summary != nil {
		post.Summary = input.Summary
	}
	if input.Cover != nil {
		post.Cover = input.Cover
	}
	if input.Content != nil {
		post.Content = *input.Content
		contentLength := utf8.RuneCountInString(post.Content)
		post.ReadTimeMinutes = uint(max(1, (contentLength+199)/200))
	}
	if input.Status != nil {
		post.Status = *input.Status
	}

	err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		var err error
		if err = s.pr.Update(ctx, post); err != nil {
			return err
		}

		if input.Tags != nil {
			if err = s.pr.SetTags(ctx, input.ID, *input.Tags); err != nil {
				return err
			}
		}

		if input.CategoryIDs != nil {
			if err = s.pr.SetCategories(ctx, input.ID, *input.CategoryIDs); err != nil {
				return err
			}
		}

		post, err = s.pr.GetAdminListItemByID(ctx, input.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost soft-deletes a post.
func (s *postService) DeletePost(ctx context.Context, user contextx.User, id uint) error {
	if err := s.authz.Authorize(ctx, user.ID, user.Role, authz.ResourcePost, authz.ActionDelete, &id); err != nil {
		return err
	}
	return s.pr.Delete(ctx, id)
}

// FlushViewCountToDB flushes accumulated view counts from Redis to the database.
func (s *postService) FlushViewCountToDB(ctx context.Context) error {
	var cursor uint64
	var allErrors []error
	s.log.Info("Flushing post view count to DB...")

	for {
		keys, next, err := s.rc.Scan(ctx, "blog:post:view_count:*", cursor, 100)
		if err != nil {
			return fmt.Errorf("failed to scan Redis keys for post view count %w", err)
		}
		cursor = next

		if len(keys) == 0 && cursor == 0 {
			break
		}
		if len(keys) == 0 {
			continue
		}

		data, err := s.rc.PopBatch(ctx, keys)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("PopBatch failed: %w", err))
		}

		updates := make(map[uint]int64)
		keysByPostID := make(map[uint]string)
		for key, valStr := range data {
			count, err := strconv.ParseInt(valStr, 10, 64)
			if err != nil || count == 0 {
				continue
			}

			parts := strings.Split(key, ":")
			if len(parts) < 4 {
				continue
			}
			postID, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
			if err != nil {
				continue
			}

			updates[uint(postID)] += count
			keysByPostID[uint(postID)] = key
		}

		if len(updates) > 0 {
			if err := s.pr.BatchIncrViewCounts(ctx, updates); err != nil {
				allErrors = append(allErrors, fmt.Errorf("db update failed: %w", err))
				allErrors = append(allErrors, restoreViewCounts(ctx, s.rc, updates, keysByPostID)...)
			}
		}

		if cursor == 0 {
			break
		}
	}
	if len(allErrors) > 0 {
		return fmt.Errorf("flush completed with %d errors: %v", len(allErrors), allErrors)
	}
	return nil
}

func restoreViewCounts(ctx context.Context, store cache.AtomicStore, updates map[uint]int64, keysByPostID map[uint]string) []error {
	var errs []error
	for postID, count := range updates {
		if _, err := store.IncrBy(ctx, keysByPostID[postID], count); err != nil {
			errs = append(errs, fmt.Errorf("restore view count for post %d: %w", postID, err))
		}
	}
	return errs
}
