package telegram

import (
	"LinkBot/storage"
	"fmt"
	"strings"
	"testing"
)

func TestFormatPagesList(t *testing.T) {
	pages := []storage.Page{
		{URL: "https://example.com/a"},
		{URL: "https://example.com/b"},
	}

	got := formatPagesList(pages)
	want := "Your saved links:\n\n1. https://example.com/a\n2. https://example.com/b"

	if len(got) != 1 {
		t.Fatalf("got %d messages, want 1", len(got))
	}
	if got[0] != want {
		t.Fatalf("got %q, want %q", got[0], want)
	}
}

func TestFormatPagesListSplitsLongMessages(t *testing.T) {
	pages := make([]storage.Page, 0, 250)
	for i := 0; i < 250; i++ {
		pages = append(pages, storage.Page{URL: fmt.Sprintf("https://example.com/%s/%d", strings.Repeat("x", 40), i)})
	}

	got := formatPagesList(pages)
	if len(got) < 2 {
		t.Fatalf("got %d messages, want at least 2", len(got))
	}

	for _, message := range got {
		if len(message) > telegramMessageLimit {
			t.Fatalf("message length is %d, limit is %d", len(message), telegramMessageLimit)
		}
	}
}
