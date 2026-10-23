package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"sort"

	bwtsearch "github.com/bgnori/textindex"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	mcpDefaultLimit   = 20
	mcpDefaultContext = 80
	mcpMaximumLimit   = 1000
	mcpMaximumContext = 1000
)

type mcpSearchInput struct {
	Pattern string `json:"pattern" jsonschema:"search pattern; FM-index uses star-free regex and other indexes use literal matching"`
	Limit   *int   `json:"limit,omitempty" jsonschema:"maximum number of matches (default 20, maximum 1000)"`
	Context *int   `json:"context,omitempty" jsonschema:"context bytes around each match (default 80, maximum 1000)"`
}

type mcpMatch struct {
	Position int    `json:"position" jsonschema:"byte offset of the match"`
	Snippet  string `json:"snippet" jsonschema:"text surrounding the match"`
}

type mcpSearchOutput struct {
	Pattern   string     `json:"pattern"`
	Matches   []mcpMatch `json:"matches"`
	Truncated bool       `json:"truncated" jsonschema:"whether additional matches were omitted"`
}

type mcpIndexInfo struct {
	IndexPath    string `json:"index_path"`
	Backend      string `json:"backend"`
	TextLength   int    `json:"text_length"`
	SuffixLength int    `json:"suffix_length,omitempty"`
	AlphabetSize int    `json:"alphabet_size,omitempty"`
}

func runMCP(args []string) error {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	indexPath := fs.String("index", "data/moby_dick.idx", "path to index file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: textindex mcp [--index FILE]")
	}

	idx, err := loadAnyIndex(*indexPath)
	if err != nil {
		return err
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "textindex", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search the loaded text index and return match positions with surrounding context.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input mcpSearchInput) (*mcp.CallToolResult, mcpSearchOutput, error) {
		output, err := searchMCPIndex(idx, input)
		return nil, output, err
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "index_info",
		Description: "Return metadata about the loaded text index.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, mcpIndexInfo, error) {
		return nil, describeMCPIndex(idx, *indexPath), nil
	})

	return server.Run(context.Background(), &mcp.StdioTransport{})
}

func searchMCPIndex(idx anyIndex, input mcpSearchInput) (mcpSearchOutput, error) {
	if input.Pattern == "" {
		return mcpSearchOutput{}, fmt.Errorf("pattern is required")
	}
	limit := mcpDefaultLimit
	if input.Limit != nil {
		limit = *input.Limit
	}
	if limit < 1 || limit > mcpMaximumLimit {
		return mcpSearchOutput{}, fmt.Errorf("limit must be between 1 and %d", mcpMaximumLimit)
	}
	contextSize := mcpDefaultContext
	if input.Context != nil {
		contextSize = *input.Context
	}
	if contextSize < 0 || contextSize > mcpMaximumContext {
		return mcpSearchOutput{}, fmt.Errorf("context must be between 0 and %d", mcpMaximumContext)
	}

	positions, _, err := searchAny(idx, input.Pattern, limit+1)
	if err != nil {
		return mcpSearchOutput{}, err
	}
	sort.Ints(positions)
	truncated := len(positions) > limit
	if truncated {
		positions = positions[:limit]
	}

	matches := make([]mcpMatch, 0, len(positions))
	for _, pos := range positions {
		matches = append(matches, mcpMatch{
			Position: pos,
			Snippet:  idx.ContextAround(pos, len(input.Pattern), contextSize),
		})
	}
	return mcpSearchOutput{Pattern: input.Pattern, Matches: matches, Truncated: truncated}, nil
}

func describeMCPIndex(idx anyIndex, indexPath string) mcpIndexInfo {
	info := mcpIndexInfo{IndexPath: indexPath, TextLength: idx.TextLen()}
	switch typed := idx.(type) {
	case *bwtsearch.Index:
		info.Backend = "fm-index"
		info.SuffixLength = typed.SALen()
		info.AlphabetSize = typed.AlphabetSize()
	case *bwtsearch.StdlibIndex:
		info.Backend = "suffixarray"
	case *bwtsearch.BiIndex:
		info.Backend = "bidirectional-fm-index"
	default:
		info.Backend = "unknown"
	}
	return info
}
