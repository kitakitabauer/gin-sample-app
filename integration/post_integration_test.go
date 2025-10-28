package integration

import (
	"context"
	"testing"

	"github.com/kitakitabauer/gin-sample-app/model"
	"github.com/kitakitabauer/gin-sample-app/repository"
	"github.com/kitakitabauer/gin-sample-app/service"
)

func newIntegrationService() *service.PostService {
	repo := repository.NewInMemoryPostRepository()
	return service.NewPostService(repo)
}

func TestPostLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	svc := newIntegrationService()

	original := model.Post{
		Title:   "First Post",
		Content: "Hello World",
		Author:  "Alice",
	}

	created, err := svc.Create(ctx, original.Title, original.Content, original.Author)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected created post to have an ID")
	}

	posts, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	fetched, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != original.Title || fetched.Author != original.Author || fetched.Content != original.Content {
		t.Fatalf("unexpected fetched post: %+v", fetched)
	}

	newTitle := "Updated Post"
	newContent := "Updated Content"
	updated, err := svc.Update(ctx, created.ID, &newTitle, &newContent, nil)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Title != newTitle || updated.Content != newContent {
		t.Fatalf("update did not apply: %+v", updated)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := svc.Get(ctx, created.ID); err != repository.ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound after delete, got %v", err)
	}
}
