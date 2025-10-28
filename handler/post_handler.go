package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kitakitabauer/gin-sample-app/internal/middleware"
	"github.com/kitakitabauer/gin-sample-app/repository"
	"github.com/kitakitabauer/gin-sample-app/service"
)

type PostHandler struct {
	service *service.PostService
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{service: service}
}

func (h *PostHandler) RegisterRoutes(router *gin.Engine) {
	posts := router.Group("/posts")
	posts.GET("", h.listPosts)
	posts.GET("/:id", h.getPost)

	protected := posts.Group("", middleware.RequireAPIKey())
	protected.POST("", h.createPost)
	protected.PATCH("/:id", h.updatePost)
	protected.DELETE("/:id", h.deletePost)
}

type createPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type updatePostRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Author  *string `json:"author"`
}

func (h *PostHandler) createPost(c *gin.Context) {
	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.service.Create(c.Request.Context(), req.Title, req.Content, req.Author)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTitleRequired),
			errors.Is(err, service.ErrContentRequired),
			errors.Is(err, service.ErrAuthorRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		}
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *PostHandler) listPosts(c *gin.Context) {
	posts, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"posts": posts})
}

func (h *PostHandler) getPost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	post, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get post"})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) updatePost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.service.Update(c.Request.Context(), id, req.Title, req.Content, req.Author)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoFieldsToUpdate),
			errors.Is(err, service.ErrTitleRequired),
			errors.Is(err, service.ErrContentRequired),
			errors.Is(err, service.ErrAuthorRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update post"})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) deletePost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		switch {
		case errors.Is(err, repository.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
