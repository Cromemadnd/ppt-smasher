package research

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ppt-smasher/internal/config"
)

type TavilySearchRequest struct {
	APIKey        string `json:"api_key"`
	Query         string `json:"query"`
	SearchDepth   string `json:"search_depth,omitempty"`
	IncludeImages bool   `json:"include_images,omitempty"`
	MaxResults    int    `json:"max_results,omitempty"`
}

type TavilySearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

type TavilyImageResult struct {
	URL string `json:"url"`
}

type TavilySearchResponse struct {
	Results []TavilySearchResult `json:"results"`
	Images  []TavilyImageResult  `json:"images,omitempty"`
}

func SearchTavily(ctx context.Context, query string, includeImages bool) (*TavilySearchResponse, error) {
	apiKey := config.GlobalConfig.Search.TavilyAPIKey
	if apiKey == "" {
		return nil, fmt.Errorf("Tavily API Key is not set")
	}

	reqBody := TavilySearchRequest{
		APIKey:        apiKey,
		Query:         query,
		SearchDepth:   "advanced",
		IncludeImages: includeImages,
		MaxResults:    5,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily search failed with status: %d", resp.StatusCode)
	}

	var searchResp TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return &searchResp, nil
}
