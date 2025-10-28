package service

import (
	"context"
	"testing"

	"github.com/kitakitabauer/gin-sample-app/repository"
)

func newTestService() *PostService {
	repo := repository.NewInMemoryPostRepository()
	return NewPostService(repo)
}

func TestPostService_Create_Success(t *testing.T) {
	svc := newTestService()

	post, err := svc.Create(context.Background(), "title", "content", "author")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if post.ID == 0 {
		t.Fatalf("expected post id to be assigned")
	}
	if post.Title != "title" || post.Content != "content" || post.Author != "author" {
		t.Fatalf("unexpected post: %+v", post)
	}
	if post.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}
}

func TestPostService_Create_ValidationErrors(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.Create(ctx, "", "content", "author"); err != ErrTitleRequired {
		t.Fatalf("expected title error, got %v", err)
	}
	if _, err := svc.Create(ctx, "title", "", "author"); err != ErrContentRequired {
		t.Fatalf("expected content error, got %v", err)
	}
	if _, err := svc.Create(ctx, "title", "content", ""); err != ErrAuthorRequired {
		t.Fatalf("expected author error, got %v", err)
	}
}

func TestPostService_ListAndGet(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	first, err := svc.Create(ctx, "first", "content", "alice")
	if err != nil {
		t.Fatalf("failed to create first post: %v", err)
	}
	second, err := svc.Create(ctx, "second", "content", "bob")
	if err != nil {
		t.Fatalf("failed to create second post: %v", err)
	}

	posts, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(posts) != 2 || posts[0].ID != first.ID || posts[1].ID != second.ID {
		t.Fatalf("unexpected posts order or length: %+v", posts)
	}

	got, err := svc.Get(ctx, first.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != first.ID {
		t.Fatalf("expected ID %d, got %d", first.ID, got.ID)
	}
}

func TestPostService_Update(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.Create(ctx, "title", "content", "author")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := svc.Update(ctx, created.ID, nil, nil, nil); err != ErrNoFieldsToUpdate {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}

	empty := ""
	if _, err := svc.Update(ctx, created.ID, &empty, nil, nil); err != ErrTitleRequired {
		t.Fatalf("expected ErrTitleRequired, got %v", err)
	}

	newTitle := "new title"
	newAuthor := "new author"
	updated, err := svc.Update(ctx, created.ID, &newTitle, nil, &newAuthor)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	if updated.Title != newTitle || updated.Author != newAuthor || updated.Content != created.Content {
		t.Fatalf("unexpected updated post: %+v", updated)
	}

	if _, err := svc.Update(ctx, created.ID+99, &newTitle, nil, nil); err != repository.ErrPostNotFound {
		t.Fatalf("expected repository.ErrPostNotFound, got %v", err)
	}
}

func TestPostService_Delete(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.Create(ctx, "title", "content", "author")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if err := svc.Delete(ctx, created.ID); err != repository.ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound after deletion, got %v", err)
	}
}
