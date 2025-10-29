package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"

	dbpkg "github.com/kitakitabauer/gin-sample-app/internal/database"
	"github.com/kitakitabauer/gin-sample-app/model"
)

func newTestSQLRepository(t *testing.T) (*SQLPostRepository, func()) {
	t.Helper()

	name := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)

	db, err := dbpkg.Open(context.Background(), dbpkg.Config{
		Driver:       "sqlite",
		DSN:          dsn,
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := dbpkg.MigrateUp(db, "sqlite"); err != nil {
		db.Close()
		t.Fatalf("failed to apply migrations: %v", err)
	}

	repo := NewSQLPostRepository(db, "sqlite")
	cleanup := func() {
		db.Close()
	}

	return repo, cleanup
}

func TestSQLPostRepository_CreateAndFind(t *testing.T) {
	repo, cleanup := newTestSQLRepository(t)
	defer cleanup()

	ctx := context.Background()
	post := model.Post{Title: "title", Content: "content", Author: "author"}

	created, err := repo.Create(ctx, post)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected ID to be set")
	}

	found, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}

	if found.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, found.ID)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll returned error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 post, got %d", len(all))
	}
}

func TestSQLPostRepository_FindByID_NotFound(t *testing.T) {
	repo, cleanup := newTestSQLRepository(t)
	defer cleanup()

	if _, err := repo.FindByID(context.Background(), 1); err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestSQLPostRepository_Update(t *testing.T) {
	repo, cleanup := newTestSQLRepository(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, model.Post{Title: "title", Content: "content", Author: "author"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	newTitle := "new title"
	newContent := "new content"
	update := PostUpdate{Title: &newTitle, Content: &newContent}

	updated, err := repo.Update(ctx, created.ID, update)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if updated.Title != newTitle || updated.Content != newContent {
		t.Fatalf("unexpected updated post: %+v", updated)
	}

	if _, err := repo.Update(ctx, created.ID+100, update); err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestSQLPostRepository_Delete(t *testing.T) {
	repo, cleanup := newTestSQLRepository(t)
	defer cleanup()

	ctx := context.Background()
	created, err := repo.Create(ctx, model.Post{Title: "title", Content: "content", Author: "author"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound after delete, got %v", err)
	}
}
