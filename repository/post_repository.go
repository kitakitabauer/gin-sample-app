package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
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

// SQLPostRepositoryはRDBを利用したPostRepositoryの実装です。
type SQLPostRepository struct {
	db      *sql.DB
	dialect string
}

func NewSQLPostRepository(db *sql.DB, driver string) *SQLPostRepository {
	return &SQLPostRepository{
		db:      db,
		dialect: detectDialect(driver),
	}
}

func (r *SQLPostRepository) Create(ctx context.Context, post model.Post) (model.Post, error) {
	switch r.dialect {
	case "postgres":
		query := `INSERT INTO posts (title, content, author, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
		if err := r.db.QueryRowContext(ctx, query, post.Title, post.Content, post.Author, post.CreatedAt).Scan(&post.ID); err != nil {
			return model.Post{}, err
		}
		return post, nil
	default:
		res, err := r.db.ExecContext(ctx, `INSERT INTO posts (title, content, author, created_at) VALUES (?, ?, ?, ?)`, post.Title, post.Content, post.Author, post.CreatedAt)
		if err != nil {
			return model.Post{}, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return model.Post{}, err
		}
		post.ID = id
		return post, nil
	}
}

func (r *SQLPostRepository) FindAll(ctx context.Context) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, content, author, created_at FROM posts ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLPostRepository) FindByID(ctx context.Context, id int64) (model.Post, error) {
	query := fmt.Sprintf(`SELECT id, title, content, author, created_at FROM posts WHERE id = %s`, r.placeholder(1))
	var post model.Post
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Post{}, ErrPostNotFound
		}
		return model.Post{}, err
	}
	return post, nil
}

func (r *SQLPostRepository) Update(ctx context.Context, id int64, update PostUpdate) (model.Post, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if update.Title != nil {
		idx := len(args) + 1
		sets = append(sets, fmt.Sprintf("title = %s", r.placeholder(idx)))
		args = append(args, *update.Title)
	}
	if update.Content != nil {
		idx := len(args) + 1
		sets = append(sets, fmt.Sprintf("content = %s", r.placeholder(idx)))
		args = append(args, *update.Content)
	}
	if update.Author != nil {
		idx := len(args) + 1
		sets = append(sets, fmt.Sprintf("author = %s", r.placeholder(idx)))
		args = append(args, *update.Author)
	}

	if len(sets) == 0 {
		return r.FindByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE posts SET %s WHERE id = %s", strings.Join(sets, ", "), r.placeholder(len(args)))

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return model.Post{}, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return model.Post{}, err
	}
	if rowsAffected == 0 {
		return model.Post{}, ErrPostNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *SQLPostRepository) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM posts WHERE id = %s", r.placeholder(1))
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *SQLPostRepository) placeholder(idx int) string {
	if r.dialect == "postgres" {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}

func detectDialect(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql", "pgx":
		return "postgres"
	default:
		return "sqlite"
	}
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
