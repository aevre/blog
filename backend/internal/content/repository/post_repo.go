// Package repository persists and queries content data.
package repository

import (
	"context"
	"time"

	"blog-server/ent"
	"blog-server/ent/post"
	"blog-server/ent/postcategory"
	"blog-server/ent/postcategoryrelation"
	"blog-server/ent/posttag"
	"blog-server/ent/posttagrelation"
	"blog-server/ent/user"
	"blog-server/internal/content/mapper"
	contentmodel "blog-server/internal/content/model"
	"blog-server/internal/platform/database"
	"blog-server/internal/shared/errx"

	"entgo.io/ent/dialect/sql"
)

// PostRepo defines persistence operations for the Post aggregate.
//
// All read operations implicitly apply:
// - soft-delete filter (DeletedAt IS NULL)
type PostRepo interface {
	Create(ctx context.Context, post *contentmodel.Post) (*contentmodel.Post, error)
	GetTagsByIDs(ctx context.Context, ids []uint) ([]contentmodel.PostTag, error)
	GetCategoriesByIDs(ctx context.Context, ids []uint) ([]contentmodel.PostCategory, error)
	Update(ctx context.Context, post *contentmodel.Post) error
	Delete(ctx context.Context, id uint) error

	ListPublished(ctx context.Context, page, pageSize int) ([]*contentmodel.Post, error)
	ListPublishedForSitemap(ctx context.Context) ([]*contentmodel.Post, error)
	ListPublishedForMeta(ctx context.Context, page, pageSize int) ([]*contentmodel.Post, error)

	ListAll(ctx context.Context, status *contentmodel.PostStatus, keyword *string, page, pageSize int) ([]*contentmodel.Post, error)
	GetAdminListItemByID(ctx context.Context, id uint) (*contentmodel.Post, error)
	GetByID(ctx context.Context, id uint) (*contentmodel.Post, error)
	GetPublishedByID(ctx context.Context, id uint) (*contentmodel.Post, error)
	PublishedVersion(ctx context.Context, id uint) (time.Time, error)
	GetLatestPublishedAt(ctx context.Context) (*time.Time, error)
	GetLatestUpdatedAt(ctx context.Context) (*time.Time, error)

	Count(ctx context.Context) (int, error)
	CountAll(ctx context.Context, status *contentmodel.PostStatus, keyword *string) (int, error)
	CountPublished(ctx context.Context) (int, error)
	CountDeleted(ctx context.Context) (int, error)

	AddTags(ctx context.Context, postID uint, tagIDs []uint) error
	SetTags(ctx context.Context, postID uint, tagIDs []uint) error
	AddCategories(ctx context.Context, postID uint, categoryIDs []uint) error
	SetCategories(ctx context.Context, postID uint, categoryIDs []uint) error

	BatchIncrViewCounts(ctx context.Context, updates map[uint]int64) error

	IsOwner(ctx context.Context, userID uint, postID uint) (bool, error)
}

type postRepo struct {
	ds *database.DataStore
}

// NewPostRepo creates a PostRepo instance using the given database.
func NewPostRepo(ds *database.DataStore) PostRepo {
	return &postRepo{ds: ds}
}

// query returns a base post query with soft-delete filter applied.
func (r *postRepo) query(ctx context.Context) *ent.PostQuery {
	return r.ds.Client(ctx).Post.Query().
		Where(post.DeletedAtIsNil())
}

// publishedQuery returns a query filtered to published posts only.
func (r *postRepo) publishedQuery(ctx context.Context) *ent.PostQuery {
	return r.query(ctx).
		Where(
			post.StatusEQ(contentmodel.PostStatusPublish),
		)
}

// deletedQuery returns a query filtered to soft-deleted posts only.
func (r *postRepo) deletedQuery(ctx context.Context) *ent.PostQuery {
	return r.ds.Client(ctx).Post.Query().
		Where(
			post.DeletedAtNotNil(),
		)
}

// Create inserts a new post record.
func (r *postRepo) Create(ctx context.Context, p *contentmodel.Post) (*contentmodel.Post, error) {
	now := time.Now()

	builder := r.ds.Client(ctx).Post.
		Create().
		SetTitle(p.Title).
		SetContent(p.Content).
		SetReadTimeMinutes(p.ReadTimeMinutes).
		SetStatus(p.Status).
		SetUserID(p.UserID).
		SetCreatedAt(now).
		SetUpdatedAt(now)

	if p.Summary != nil {
		builder.SetSummary(*p.Summary)
	}
	if p.Cover != nil {
		builder.SetCover(*p.Cover)
	}
	if p.Status == contentmodel.PostStatusPublish {
		builder.SetPublishedAt(now)
	}

	ep, err := builder.Save(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPost(ep), nil
}

// GetTagsByIDs returns tags for the given IDs.
func (r *postRepo) GetTagsByIDs(ctx context.Context, ids []uint) ([]contentmodel.PostTag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	tags, err := r.ds.Client(ctx).PostTag.
		Query().
		Where(posttag.IDIn(ids...)).
		Select(posttag.FieldID, posttag.FieldName, posttag.FieldSlug).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToPostTags(tags), nil
}

// GetCategoriesByIDs returns categories for the given IDs.
func (r *postRepo) GetCategoriesByIDs(ctx context.Context, ids []uint) ([]contentmodel.PostCategory, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	categories, err := r.ds.Client(ctx).PostCategory.
		Query().
		Where(postcategory.IDIn(ids...)).
		Select(postcategory.FieldID, postcategory.FieldName, postcategory.FieldSlug).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToPostCategories(categories), nil
}

// Update updates an existing post record.
func (r *postRepo) Update(ctx context.Context, p *contentmodel.Post) error {
	builder := r.ds.Client(ctx).Post.
		UpdateOneID(p.ID).
		SetUpdatedAt(time.Now())

	if p.Title != "" {
		builder.SetTitle(p.Title)
	}

	if p.Summary != nil {
		builder.SetSummary(*p.Summary)
	}

	if p.Cover != nil {
		builder.SetCover(*p.Cover)
	}

	if p.Content != "" {
		builder.SetContent(p.Content)
	}

	if p.Status != "" {
		builder.SetStatus(p.Status)
		if p.Status == contentmodel.PostStatusPublish {
			builder.SetPublishedAt(time.Now())
		}
	}

	if p.ReadTimeMinutes != 0 {
		builder.SetReadTimeMinutes(p.ReadTimeMinutes)
	}

	err := builder.Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return nil
}

// Delete soft-deletes a post by setting deleted_at.
func (r *postRepo) Delete(ctx context.Context, id uint) error {
	now := time.Now()
	err := r.ds.Client(ctx).Post.
		UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// ListPublished returns all published posts for list views.
func (r *postRepo) ListPublished(ctx context.Context, page, pageSize int) ([]*contentmodel.Post, error) {
	page, pageSize = normalizedPage(page, pageSize)
	ps, err := r.publishedQuery(ctx).
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldCover,
			post.FieldReadTimeMinutes,
			post.FieldViewCount,
			post.FieldPublishedAt,
			post.FieldCreatedAt,
			post.FieldUpdatedAt,
		).
		WithAuthor(func(q *ent.UserQuery) {
			q.Select(user.FieldUsername)
		}).
		WithCategories().
		WithTags().
		Order(
			post.ByPublishedAt(sql.OrderDesc()),
		).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPosts(ps), nil
}

// ListPublishedForSitemap returns minimal post data for sitemap generation.
func (r *postRepo) ListPublishedForSitemap(ctx context.Context) ([]*contentmodel.Post, error) {
	ps, err := r.publishedQuery(ctx).
		Select(
			post.FieldID,
			post.FieldUpdatedAt,
		).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPosts(ps), nil
}

// ListPublishedForMeta returns lightweight post metadata for SEO and previews.
func (r *postRepo) ListPublishedForMeta(ctx context.Context, page, pageSize int) ([]*contentmodel.Post, error) {
	page, pageSize = normalizedPage(page, pageSize)

	ps, err := r.publishedQuery(ctx).
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldUpdatedAt,
			post.FieldPublishedAt,
		).
		Order(
			post.ByPublishedAt(sql.OrderDesc()),
		).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPosts(ps), nil
}

// ListAll returns all posts including drafts for admin list views.
func (r *postRepo) ListAll(ctx context.Context, status *contentmodel.PostStatus, keyword *string, page, pageSize int) ([]*contentmodel.Post, error) {
	page, pageSize = normalizedPage(page, pageSize)

	query := r.query(ctx)

	if status != nil {
		query = query.Where(post.StatusEQ(*status))
	}

	if keyword != nil && *keyword != "" {
		query = query.Where(post.TitleContainsFold(*keyword))
	}

	ps, err := query.
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldCover,
			post.FieldStatus,
			post.FieldReadTimeMinutes,
			post.FieldViewCount,
			post.FieldPublishedAt,
			post.FieldCreatedAt,
			post.FieldUpdatedAt,
		).
		WithAuthor(func(q *ent.UserQuery) {
			q.Select(user.FieldUsername)
		}).
		WithCategories().
		WithTags().
		Order(
			post.ByCreatedAt(sql.OrderDesc()),
		).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPosts(ps), nil
}

// GetAdminListItemByID returns the same lightweight shape used by the admin list.
func (r *postRepo) GetAdminListItemByID(ctx context.Context, id uint) (*contentmodel.Post, error) {
	p, err := r.query(ctx).
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldCover,
			post.FieldStatus,
			post.FieldReadTimeMinutes,
			post.FieldViewCount,
			post.FieldPublishedAt,
			post.FieldCreatedAt,
			post.FieldUpdatedAt,
		).
		WithAuthor(func(q *ent.UserQuery) {
			q.Select(user.FieldUsername)
		}).
		WithCategories().
		WithTags().
		Where(
			post.IDEQ(id),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPost(p), nil
}

// CountAll returns the count of posts matching optional filters.
func (r *postRepo) CountAll(ctx context.Context, status *contentmodel.PostStatus, keyword *string) (int, error) {
	query := r.query(ctx)

	if status != nil {
		query = query.Where(post.StatusEQ(contentmodel.PostStatus(*status)))
	}

	if keyword != nil && *keyword != "" {
		query = query.Where(post.TitleContainsFold(*keyword))
	}

	return query.Count(ctx)
}

// PublishedVersion checks visibility even when the detail is cached.
func (r *postRepo) PublishedVersion(ctx context.Context, id uint) (time.Time, error) {
	p, err := r.publishedQuery(ctx).Where(post.IDEQ(id)).Select(post.FieldUpdatedAt).First(ctx)
	if ent.IsNotFound(err) {
		return time.Time{}, errx.New(errx.CodeNotFound, err)
	}
	if err != nil {
		return time.Time{}, errx.New(errx.CodeInternalError, err)
	}
	return p.UpdatedAt, nil
}

// GetPublishedByID returns a single published post by ID.
func (r *postRepo) GetPublishedByID(ctx context.Context, id uint) (*contentmodel.Post, error) {
	p, err := r.publishedQuery(ctx).
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldCover,
			post.FieldContent,
			post.FieldReadTimeMinutes,
			post.FieldViewCount,
			post.FieldPublishedAt,
			post.FieldCreatedAt,
			post.FieldUpdatedAt,
		).
		WithAuthor(func(q *ent.UserQuery) {
			q.Select(user.FieldUsername)
		}).
		WithCategories().
		WithTags().
		Where(
			post.IDEQ(id),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPost(p), nil
}

// GetByID returns a single post by ID.
func (r *postRepo) GetByID(ctx context.Context, id uint) (*contentmodel.Post, error) {
	p, err := r.query(ctx).
		Select(
			post.FieldID,
			post.FieldTitle,
			post.FieldSummary,
			post.FieldCover,
			post.FieldContent,
			post.FieldStatus,
			post.FieldReadTimeMinutes,
			post.FieldViewCount,
			post.FieldPublishedAt,
			post.FieldCreatedAt,
			post.FieldUpdatedAt,
		).
		WithAuthor(func(q *ent.UserQuery) {
			q.Select(user.FieldUsername)
		}).
		WithCategories().
		WithTags().
		Where(
			post.IDEQ(id),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToPost(p), nil
}

// GetLatestPublishedAt returns the most recent published timestamp.
func (r *postRepo) GetLatestPublishedAt(ctx context.Context) (*time.Time, error) {
	p, err := r.publishedQuery(ctx).
		Select(post.FieldPublishedAt).
		Order(
			post.ByPublishedAt(sql.OrderDesc()),
		).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return p.PublishedAt, nil
}

// GetLatestUpdatedAt returns the most recent update timestamp.
func (r *postRepo) GetLatestUpdatedAt(ctx context.Context) (*time.Time, error) {
	p, err := r.publishedQuery(ctx).
		Select(post.FieldUpdatedAt).
		Order(
			post.ByUpdatedAt(sql.OrderDesc()),
		).
		First(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return &p.UpdatedAt, nil
}

// Count returns the total number of posts (including deleted).
func (r *postRepo) Count(ctx context.Context) (int, error) {
	return r.query(ctx).Count(ctx)
}

// CountPublished returns the number of published posts.
func (r *postRepo) CountPublished(ctx context.Context) (int, error) {
	return r.publishedQuery(ctx).Count(ctx)
}

// CountDeleted returns the number of soft-deleted posts.
func (r *postRepo) CountDeleted(ctx context.Context) (int, error) {
	return r.deletedQuery(ctx).Count(ctx)
}

// AddTags attaches tags to a post without removing existing relations.
func (r *postRepo) AddTags(ctx context.Context, postID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}

	err := r.ds.Client(ctx).
		PostTagRelation.
		MapCreateBulk(
			tagIDs,
			func(c *ent.PostTagRelationCreate, i int) {
				c.
					SetPostID(postID).
					SetPostTagID(tagIDs[i])
			},
		).
		OnConflict(
			sql.ConflictColumns(
				posttagrelation.FieldPostID,
				posttagrelation.FieldPostTagID,
			),
		).
		DoNothing().
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// SetTags replaces all tags of a post (clear then add).
func (r *postRepo) SetTags(ctx context.Context, postID uint, tagIDs []uint) error {
	_, err := r.ds.Client(ctx).
		PostTagRelation.
		Delete().
		Where(posttagrelation.PostIDEQ(postID)).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return r.AddTags(ctx, postID, tagIDs)
}

// AddCategories attaches categories to a post without removing existing relations.
func (r *postRepo) AddCategories(ctx context.Context, postID uint, categoryIDs []uint) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	err := r.ds.Client(ctx).
		PostCategoryRelation.
		MapCreateBulk(
			categoryIDs,
			func(c *ent.PostCategoryRelationCreate, i int) {
				c.
					SetPostID(postID).
					SetPostCategoryID(categoryIDs[i])
			},
		).
		OnConflict(
			sql.ConflictColumns(
				postcategoryrelation.FieldPostID,
				postcategoryrelation.FieldPostCategoryID,
			),
		).
		DoNothing().
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// SetCategories replaces all categories of a post (clear then add).
func (r *postRepo) SetCategories(ctx context.Context, postID uint, categoryIDs []uint) error {
	_, err := r.ds.Client(ctx).
		PostCategoryRelation.
		Delete().
		Where(postcategoryrelation.PostIDEQ(postID)).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return r.AddCategories(ctx, postID, categoryIDs)
}

// BatchIncrViewCounts performs a bulk increment of view counts.
//
// Note: query complexity grows linearly with batch size.
func (r *postRepo) BatchIncrViewCounts(ctx context.Context, updates map[uint]int64) error {
	for id, delta := range updates {
		if delta <= 0 {
			delete(updates, id)
		}
	}
	if len(updates) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(updates))
	for id := range updates {
		ids = append(ids, id)
	}

	err := r.ds.Client(ctx).Post.
		Update().
		Where(post.IDIn(ids...)).
		Modify(func(u *sql.UpdateBuilder) {
			caseExpr := func(b *sql.Builder) {
				b.WriteString("CASE ")
				b.Ident(post.FieldID)
				b.WriteString(" ")

				for _, id := range ids {
					b.WriteString("WHEN ")
					b.Arg(id)
					b.WriteString(" THEN ")
					b.Ident(post.FieldViewCount)
					b.WriteString(" + ")
					b.Arg(updates[id])
					b.WriteString(" ")
				}

				b.WriteString("ELSE ")
				b.Ident(post.FieldViewCount)
				b.WriteString(" END")
			}
			u.Set(post.FieldViewCount, sql.ExprFunc(caseExpr))
		}).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}

// IsOwner checks whether a user is the author of a given post.
func (r *postRepo) IsOwner(ctx context.Context, userID uint, postID uint) (bool, error) {
	exists, err := r.query(ctx).
		Where(
			post.IDEQ(postID),
			post.UserIDEQ(userID),
		).
		Exist(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}
