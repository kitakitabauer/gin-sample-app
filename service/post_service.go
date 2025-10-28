package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kitakitabauer/gin-sample-app/model"
	"github.com/kitakitabauer/gin-sample-app/repository"
)

var (
	ErrTitleRequired    = errors.New("title is required")
	ErrContentRequired  = errors.New("content is required")
	ErrAuthorRequired   = errors.New("author is required")
	ErrNoFieldsToUpdate = errors.New("no fields provided to update")
)

type PostService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) Create(ctx context.Context, title, content, author string) (model.Post, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	author = strings.TrimSpace(author)

	if title == "" {
		return model.Post{}, ErrTitleRequired
	}
	if content == "" {
		return model.Post{}, ErrContentRequired
	}
	if author == "" {
		return model.Post{}, ErrAuthorRequired
	}

	post := model.Post{
		Title:     title,
		Content:   content,
		Author:    author,
		CreatedAt: time.Now().UTC(),
	}

	created, err := s.repo.Create(ctx, post)
	if err != nil {
		return model.Post{}, err
	}

	return created, nil
}

func (s *PostService) List(ctx context.Context) ([]model.Post, error) {
	return s.repo.FindAll(ctx)
}

func (s *PostService) Get(ctx context.Context, id int64) (model.Post, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *PostService) Update(ctx context.Context, id int64, title, content, author *string) (model.Post, error) {
	var update repository.PostUpdate
	var hasUpdate bool

	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return model.Post{}, ErrTitleRequired
		}
		update.Title = &trimmed
		hasUpdate = true
	}

	if content != nil {
		trimmed := strings.TrimSpace(*content)
		if trimmed == "" {
			return model.Post{}, ErrContentRequired
		}
		update.Content = &trimmed
		hasUpdate = true
	}

	if author != nil {
		trimmed := strings.TrimSpace(*author)
		if trimmed == "" {
			return model.Post{}, ErrAuthorRequired
		}
		update.Author = &trimmed
		hasUpdate = true
	}

	if !hasUpdate {
		return model.Post{}, ErrNoFieldsToUpdate
	}

	return s.repo.Update(ctx, id, update)
}

func (s *PostService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
