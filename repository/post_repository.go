package repository

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/kitakitabauer/gin-sample-app/model"
)

var ErrPostNotFound = errors.New("post not found")

// PostRepositoryはPostの永続化を抽象化するインターフェースです。
type PostRepository interface {
	Create(ctx context.Context, post model.Post) (model.Post, error)
	FindAll(ctx context.Context) ([]model.Post, error)
	FindByID(ctx context.Context, id int64) (model.Post, error)
	Update(ctx context.Context, id int64, update PostUpdate) (model.Post, error)
	Delete(ctx context.Context, id int64) error
}

type PostUpdate struct {
	Title   *string
	Content *string
	Author  *string
}

// InMemoryPostRepositoryはPostRepositoryのメモリ上の実装です。
type InMemoryPostRepository struct {
	mu     sync.RWMutex
	posts  map[int64]model.Post
	nextID int64
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
	return &InMemoryPostRepository{
		posts:  make(map[int64]model.Post),
		nextID: 0,
	}
}

func (r *InMemoryPostRepository) Create(_ context.Context, post model.Post) (model.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	post.ID = r.nextID
	r.posts[post.ID] = post

	return post, nil
}

func (r *InMemoryPostRepository) FindAll(_ context.Context) ([]model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Post, 0, len(r.posts))
	for _, post := range r.posts {
		result = append(result, post)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (r *InMemoryPostRepository) FindByID(_ context.Context, id int64) (model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.posts[id]
	if !ok {
		return model.Post{}, ErrPostNotFound
	}

	return post, nil
}

func (r *InMemoryPostRepository) Update(_ context.Context, id int64, update PostUpdate) (model.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, ok := r.posts[id]
	if !ok {
		return model.Post{}, ErrPostNotFound
	}

	if update.Title != nil {
		post.Title = *update.Title
	}
	if update.Content != nil {
		post.Content = *update.Content
	}
	if update.Author != nil {
		post.Author = *update.Author
	}

	r.posts[id] = post
	return post, nil
}

func (r *InMemoryPostRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[id]; !ok {
		return ErrPostNotFound
	}

	delete(r.posts, id)
	return nil
}
