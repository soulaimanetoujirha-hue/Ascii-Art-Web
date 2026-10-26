package ascii

import (
	"errors"
	"strings"
	"testing"
)

// Tests run inside the ascii folder, so the banners are one folder up.
const standard = "../banners/standard.txt"

func TestRenderHi(t *testing.T) {
	want := strings.Join([]string{
		" _    _   _  ",
		"| |  | | (_) ",
		"| |__| |  _  ",
		"|  __  | | | ",
		"| |  | | | | ",
		"|_|  |_| |_| ",
		"             ",
		"             ",
	}, "\n") + "\n"

	got, err := Render("Hi", standard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("wrong art:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderNewlines(t *testing.T) {
	// "a\r\n\r\nb": a (8 rows), the empty line (1 row), b (8 rows)
	got, err := Render("a\r\n\r\nb", standard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rows := strings.Count(got, "\n"); rows != 17 {
		t.Errorf("got %d rows, want 17", rows)
	}
}

func TestRenderInvalidChar(t *testing.T) {
	_, err := Render("héllo", standard)
	if !errors.Is(err, ErrInvalidChar) {
		t.Errorf("got %v, want ErrInvalidChar", err)
	}
}

func TestRenderMissingBanner(t *testing.T) {
	_, err := Render("hi", "../banners/nope.txt")
	if !errors.Is(err, ErrBannerNotFound) {
		t.Errorf("got %v, want ErrBannerNotFound", err)
	}
}
