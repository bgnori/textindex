package main

import (
	"testing"

	bwtsearch "github.com/bgnori/textindex"
)

func TestSearchMCPIndex(t *testing.T) {
	idx := bwtsearch.Build([]byte("abracadabra abracadabra"))
	limit, contextSize := 2, 2

	got, err := searchMCPIndex(idx, mcpSearchInput{Pattern: "abra", Limit: &limit, Context: &contextSize})
	if err != nil {
		t.Fatalf("searchMCPIndex failed: %v", err)
	}
	if len(got.Matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(got.Matches))
	}
	validPositions := map[int]bool{0: true, 7: true, 12: true, 19: true}
	previous := -1
	for _, match := range got.Matches {
		if !validPositions[match.Position] || match.Position <= previous {
			t.Fatalf("match positions are invalid or unsorted: %v", []int{got.Matches[0].Position, got.Matches[1].Position})
		}
		previous = match.Position
	}
	if !got.Truncated {
		t.Fatal("expected truncated results")
	}
}

func TestSearchMCPIndexRejectsInvalidLimits(t *testing.T) {
	idx := bwtsearch.Build([]byte("hello"))
	negative, tooMany := -1, mcpMaximumLimit+1
	negativeContext, tooMuchContext := -1, mcpMaximumContext+1
	for _, input := range []mcpSearchInput{
		{Pattern: "hello", Limit: &negative},
		{Pattern: "hello", Limit: &tooMany},
		{Pattern: "hello", Context: &negativeContext},
		{Pattern: "hello", Context: &tooMuchContext},
		{},
	} {
		if _, err := searchMCPIndex(idx, input); err == nil {
			t.Fatalf("searchMCPIndex(%+v) expected an error", input)
		}
	}
}

func TestSearchMCPIndexAllowsZeroContext(t *testing.T) {
	idx := bwtsearch.Build([]byte("hello hello"))
	contextSize := 0
	got, err := searchMCPIndex(idx, mcpSearchInput{Pattern: "hello", Context: &contextSize})
	if err != nil {
		t.Fatalf("searchMCPIndex failed: %v", err)
	}
	if got.Matches[0].Snippet != "hello" {
		t.Fatalf("snippet = %q, want %q", got.Matches[0].Snippet, "hello")
	}
}
