package sqlite

import (
	"LinkBot/storage"
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestStorageListAndRemove(t *testing.T) {
	ctx := context.Background()

	s, err := New(filepath.Join(t.TempDir(), "nested", "storage.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := s.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	pages := []*storage.Page{
		{URL: "https://example.com/b", UserName: "alice"},
		{URL: "https://example.com/a", UserName: "alice"},
		{URL: "https://example.com/other", UserName: "bob"},
	}

	for _, page := range pages {
		if err := s.Save(ctx, page); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}
	if err := s.Save(ctx, pages[0]); err != nil {
		t.Fatalf("duplicate Save() error = %v", err)
	}

	got, err := s.List(ctx, "alice")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d pages, want 2", len(got))
	}
	if got[0].URL != "https://example.com/a" || got[1].URL != "https://example.com/b" {
		t.Fatalf("got pages %#v in unexpected order", got)
	}

	if err := s.Remove(ctx, &got[0]); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	exists, err := s.Exists(ctx, &got[0])
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Fatal("removed page still exists")
	}
}

func TestStorageListNoSavedPages(t *testing.T) {
	ctx := context.Background()

	s, err := New(filepath.Join(t.TempDir(), "storage.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := s.Init(ctx); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	_, err = s.List(ctx, "alice")
	if !errors.Is(err, storage.ErrNoSavedPages) {
		t.Fatalf("List() error = %v, want %v", err, storage.ErrNoSavedPages)
	}
}
