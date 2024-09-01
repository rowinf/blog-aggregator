package main

import (
	"testing"
	"time"
)

func TestFetchRSSFeed(t *testing.T) {
	url := "https://blog.boot.dev/index.xml"
	response := FetchRSSFeed(url)
	want := "Boot.dev Blog"
	if want != response.Channel.Title {
		t.Fatalf(`UpdateFeed(url) = %q, want match for %#q, nil`, want, url)
	}
}

func TestParsePubDate(t *testing.T) {
	have := "Fri, 26 Jul 2024 00:00:00 +0000"
	want := time.Date(2024, time.July, 26, 0, 0, 0, 0, time.UTC)

	parsedTime, err := ParseDate(have)
	if err != nil {
		t.Fatalf("ParseDate returned an error: %v", err)
	}

	if !parsedTime.Equal(want) {
		t.Errorf("ParseDate returned %v, want %v", parsedTime, want)
	}
	invalidDateStr := "invalid date"
	_, err = ParseDate(invalidDateStr)
	if err == nil {
		t.Error("ParseDate should have returned an error for an invalid date string")
	}
}
