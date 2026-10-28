package handler

import (
	"fmt"
	"strconv"

	"blog-server/contextx"
	"blog-server/entity"
	"blog-server/middleware"
	"blog-server/pkg/errx"
	"blog-server/pkg/validatorx"
	"blog-server/request"
	"blog-server/response"
	"blog-server/service"

	"github.com/labstack/echo/v5"
)

// PostHandler defines the interface for post HTTP handlers.
type PostHandler interface {
	GetPosts(c *echo.Context) error
	GetPost(c *echo.Context) error
	GetPostIds(c *echo.Context) error
	CreatePost(c *echo.Context) error

	AdminGetPosts(c *echo.Context) error
	AdminGetPost(c *echo.Context) error
	UpdatePost(c *echo.Context) error
	DeletePost(c *echo.Context) error
}

// postHandler implements the PostHandler interface.
type postHandler struct {
	svc      service.PostService
	validate validatorx.Validator
}

// NewPostHandler creates a new post handler instance.
func NewPostHandler(svc service.PostService, validate validatorx.Validator) PostHandler {
	return &postHandler{svc: svc, validate: validate}
}

// GetPosts retrieves all published posts.
func (h *postHandler) GetPosts(c *echo.Context) error {
	query := &request.PostPageReq{Page: 1, PageSize: 10}
	if err := c.Bind(query); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(query); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	posts, total, err := h.svc.GetPosts(c.Request().Context(), query.Page, query.PageSize)
	if err != nil {
		return err
	}

	postDTOs := make([]response.PostListRes, len(posts))
	for i, post := range posts {
		postDTOs[i] = toPostListRes(post)
	}

	return response.OK(c, response.Success(response.Page[response.PostListRes]{
		Total: total,
		List:  postDTOs,
	}))
}

// GetPost retrieves a single published post by ID.
func (h *postHandler) GetPost(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	post, err := h.svc.GetPostByID(c.Request().Context(), uint(id))
	if err != nil {
		return err
	}

	return response.OK(c, response.Success(toPostRes(post)))
}

// GetPostIds retrieves metadata (id and updated_at) for all posts.
func (h *postHandler) GetPostIds(c *echo.Context) error {
	metas := h.svc.GetPostsMeta(c.Request().Context())

	metasDTO := make([]response.PostMetaRes, len(metas))
	for i, meta := range metas {
		metasDTO[i] = response.PostMetaRes{
			ID:        meta.ID,
			UpdatedAt: meta.UpdatedAt,
		}
	}

	return response.OK(c, response.Success(metasDTO))
}

// CreatePost creates a new post.
func (h *postHandler) CreatePost(c *echo.Context) error {
	req := new(request.CreatePostReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	u, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}

	input := &service.CreatePostInput{
		Title:       req.Title,
		Summary:     req.Summary,
		Cover:       req.Cover,
		Content:     req.Content,
		UserID:      u.ID,
		Status:      req.Status,
		Tags:        req.Tags,
		CategoryIDs: req.CategoryIDs,
	}

	post, err := h.svc.CreatePost(c.Request().Context(), u, input)
	if err != nil {
		return err
	}

	return response.OK(c, response.Success(toAdminPostListRes(post)))
}

// AdminGetPosts retrieves all posts (including drafts) for admin.
func (h *postHandler) AdminGetPosts(c *echo.Context) error {
	query := new(request.AdminPostListReq)
	if err := c.Bind(query); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(query); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	u, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}
	posts, total, err := h.svc.AdminGetPosts(c.Request().Context(), u, query.Status, query.Keyword, query.Page, query.PageSize)
	if err != nil {
		return err
	}

	postDTOs := make([]response.AdminPostListRes, len(posts))
	for i, post := range posts {
		postDTOs[i] = toAdminPostListRes(post)
	}

	return response.OK(c, response.Success(response.Page[response.AdminPostListRes]{
		Total: total,
		List:  postDTOs,
	}))
}

// AdminGetPost retrieves a single post by ID for admin.
func (h *postHandler) AdminGetPost(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	u, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}
	post, err := h.svc.AdminGetPostByID(c.Request().Context(), u, uint(id))
	if err != nil {
		return err
	}

	return response.OK(c, response.Success(toAdminPostRes(post)))
}

// UpdatePost updates a post.
func (h *postHandler) UpdatePost(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	req := new(request.UpdatePostReq)
	if err := c.Bind(req); err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	if err := h.validate.Struct(req); err != nil {
		return errx.New(errx.CodeValidationFailed, err)
	}

	u, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}

	input := &service.UpdatePostInput{
		ID:          uint(id),
		Title:       req.Title,
		Summary:     req.Summary,
		Cover:       req.Cover,
		Content:     req.Content,
		Status:      req.Status,
		Tags:        req.Tags,
		CategoryIDs: req.CategoryIDs,
	}

	post, err := h.svc.UpdatePost(c.Request().Context(), u, input)
	if err != nil {
		return err
	}

	return response.OK(c, response.Success(toAdminPostListRes(post)))
}

// DeletePost deletes a post.
func (h *postHandler) DeletePost(c *echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errx.New(errx.CodeInvalidParam, err)
	}

	u, ok := contextx.GetUser(c.Request().Context())
	if !ok {
		return errx.New(errx.CodeUnauthorized, fmt.Errorf("missing user in context"))
	}

	if err := h.svc.DeletePost(c.Request().Context(), u, uint(id)); err != nil {
		return err
	}

	return response.OK(c, response.Success[any](nil))
}

// RegisterPostRoutes registers all post-related routes.
func RegisterPostRoutes(r *echo.Group, h PostHandler, am *middleware.AuthMiddleware) {
	group := r.Group("/posts")
	group.GET("", h.GetPosts)
	group.GET("/meta", h.GetPostIds)
	group.GET("/:id", h.GetPost)
	group.POST("", h.CreatePost, am.Handler())

	// Admin routes
	adminGroup := r.Group("/admin/posts")
	adminGroup.GET("", h.AdminGetPosts, am.Handler())
	adminGroup.GET("/:id", h.AdminGetPost, am.Handler())
	adminGroup.PUT("/:id", h.UpdatePost, am.Handler())
	adminGroup.DELETE("/:id", h.DeletePost, am.Handler())
}

// toPostRes maps a domain Post to the detail response DTO.
func toPostRes(p *entity.Post) response.PostRes {
	return response.PostRes{
		ID:              p.ID,
		Title:           p.Title,
		Summary:         p.Summary,
		Content:         p.Content,
		Cover:           p.Cover,
		ReadTimeMinutes: p.ReadTimeMinutes,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		UpdatedAt:       p.UpdatedAt,
		Author:          p.User,
		Tags:            toTagResList(p.Tags),
		Categories:      toCategoryResList(p.Categories),
	}
}

// toPostListRes maps a domain Post to the list response DTO (no content).
func toPostListRes(p *entity.Post) response.PostListRes {
	return response.PostListRes{
		ID:              p.ID,
		Title:           p.Title,
		Summary:         p.Summary,
		Cover:           p.Cover,
		ReadTimeMinutes: p.ReadTimeMinutes,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		UpdatedAt:       p.UpdatedAt,
		Author:          p.User,
		Tags:            toTagResList(p.Tags),
		Categories:      toCategoryResList(p.Categories),
	}
}

// toAdminPostRes maps a domain Post to the admin detail response DTO.
func toAdminPostRes(p *entity.Post) response.AdminPostRes {
	return response.AdminPostRes{
		ID:              p.ID,
		Title:           p.Title,
		Summary:         p.Summary,
		Content:         p.Content,
		Cover:           p.Cover,
		Status:          string(p.Status),
		ReadTimeMinutes: p.ReadTimeMinutes,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Author:          p.User,
		Tags:            toTagResList(p.Tags),
		Categories:      toCategoryResList(p.Categories),
	}
}

// toAdminPostListRes maps a domain Post to the admin list response DTO.
func toAdminPostListRes(p *entity.Post) response.AdminPostListRes {
	return response.AdminPostListRes{
		ID:              p.ID,
		Title:           p.Title,
		Summary:         p.Summary,
		Cover:           p.Cover,
		Status:          string(p.Status),
		ReadTimeMinutes: p.ReadTimeMinutes,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Author:          p.User,
		Tags:            toTagResList(p.Tags),
		Categories:      toCategoryResList(p.Categories),
	}
}

// toTagResList converts entity tags to response DTOs.
func toTagResList(tags []entity.PostTag) []response.PostTagRes {
	if len(tags) == 0 {
		return []response.PostTagRes{}
	}
	result := make([]response.PostTagRes, len(tags))
	for i, tag := range tags {
		result[i] = response.PostTagRes{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		}
	}
	return result
}

// toCategoryResList converts entity categories to response DTOs.
func toCategoryResList(categories []entity.PostCategory) []response.PostCategoryRes {
	if len(categories) == 0 {
		return []response.PostCategoryRes{}
	}
	result := make([]response.PostCategoryRes, len(categories))
	for i, cat := range categories {
		result[i] = response.PostCategoryRes{
			ID:   cat.ID,
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}
	return result
}
