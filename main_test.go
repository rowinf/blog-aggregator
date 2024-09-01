package main

import "testing"

func TestFetchRSSFeed(t *testing.T) {
	url := "https://blog.boot.dev/index.xml"
	response := FetchRSSFeed(url)
	want := "Boot.dev Blog"
	if want != response.Channel.Title {
		t.Fatalf(`UpdateFeed(url) = %q, want match for %#q, nil`, want, url)
	}
}
