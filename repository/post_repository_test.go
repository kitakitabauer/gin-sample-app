package repository

import (
	"context"
	"testing"

	"github.com/kitakitabauer/gin-sample-app/model"
)

func TestInMemoryPostRepository_CreateAndFind(t *testing.T) {
	repo := NewInMemoryPostRepository()
	ctx := context.Background()

	post := model.Post{
		Title:   "title",
		Content: "content",
		Author:  "author",
	}

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

func TestInMemoryPostRepository_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryPostRepository()

	if _, err := repo.FindByID(context.Background(), 1); err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestInMemoryPostRepository_Update(t *testing.T) {
	repo := NewInMemoryPostRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, model.Post{
		Title:   "title",
		Content: "content",
		Author:  "author",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	newTitle := "new title"
	newContent := "new content"
	update := PostUpdate{
		Title:   &newTitle,
		Content: &newContent,
	}

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

func TestInMemoryPostRepository_Delete(t *testing.T) {
	repo := NewInMemoryPostRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, model.Post{
		Title:   "title",
		Content: "content",
		Author:  "author",
	})
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
