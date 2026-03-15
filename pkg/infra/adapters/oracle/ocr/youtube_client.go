package oracle_ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// YouTubeClient discovers live streams for CS2 tournaments using the YouTube Data API v3
type YouTubeClient struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// LiveStream represents a discovered YouTube live stream
type LiveStream struct {
	VideoID      string    `json:"video_id"`
	Title        string    `json:"title"`
	ChannelID    string    `json:"channel_id"`
	ChannelTitle string    `json:"channel_title"`
	StreamURL    string    `json:"stream_url"`
	ViewerCount  int64     `json:"viewer_count,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
}

// NewYouTubeClient creates a YouTube Data API v3 client
func NewYouTubeClient(apiKey string) *YouTubeClient {
	return &YouTubeClient{
		apiKey:  apiKey,
		baseURL: "https://www.googleapis.com/youtube/v3",
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SearchLiveStreams searches YouTube for live CS2 tournament streams
func (y *YouTubeClient) SearchLiveStreams(ctx context.Context, query string, maxResults int) ([]LiveStream, error) {
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 10
	}

	params := url.Values{
		"part":       {"snippet"},
		"q":          {query},
		"type":       {"video"},
		"eventType":  {"live"},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
		"order":      {"viewCount"},
		"key":        {y.apiKey},
	}

	reqURL := fmt.Sprintf("%s/search?%s", y.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("youtube API error %d: %s", resp.StatusCode, string(body))
	}

	var searchResp youtubeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	streams := make([]LiveStream, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		stream := LiveStream{
			VideoID:      item.ID.VideoID,
			Title:        item.Snippet.Title,
			ChannelID:    item.Snippet.ChannelID,
			ChannelTitle: item.Snippet.ChannelTitle,
			StreamURL:    fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID),
		}
		streams = append(streams, stream)
	}

	slog.InfoContext(ctx, "youtube live stream search completed",
		slog.String("query", query),
		slog.Int("results", len(streams)),
	)

	return streams, nil
}

// GetVideoDetails fetches detailed info about a specific video (including live streaming details)
func (y *YouTubeClient) GetVideoDetails(ctx context.Context, videoID string) (*LiveStream, error) {
	params := url.Values{
		"part": {"snippet,liveStreamingDetails"},
		"id":   {videoID},
		"key":  {y.apiKey},
	}

	reqURL := fmt.Sprintf("%s/videos?%s", y.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("video details request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("youtube API error %d: %s", resp.StatusCode, string(body))
	}

	var videoResp youtubeVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&videoResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(videoResp.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	item := videoResp.Items[0]
	stream := &LiveStream{
		VideoID:      videoID,
		Title:        item.Snippet.Title,
		ChannelID:    item.Snippet.ChannelID,
		ChannelTitle: item.Snippet.ChannelTitle,
		StreamURL:    fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	}

	if item.LiveStreamingDetails != nil {
		if item.LiveStreamingDetails.ConcurrentViewers != "" {
			fmt.Sscanf(item.LiveStreamingDetails.ConcurrentViewers, "%d", &stream.ViewerCount)
		}
		if item.LiveStreamingDetails.ActualStartTime != "" {
			t, _ := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
			stream.StartedAt = t
		}
	}

	return stream, nil
}

// IsVideoLive checks if a specific YouTube video is currently live
func (y *YouTubeClient) IsVideoLive(ctx context.Context, videoID string) (bool, error) {
	stream, err := y.GetVideoDetails(ctx, videoID)
	if err != nil {
		return false, err
	}

	// If we got live streaming details with start time and no end time, it's live
	return !stream.StartedAt.IsZero(), nil
}

// --- YouTube API response types ---

type youtubeSearchResponse struct {
	Items []youtubeSearchItem `json:"items"`
}

type youtubeSearchItem struct {
	ID      youtubeSearchItemID `json:"id"`
	Snippet youtubeSnippet      `json:"snippet"`
}

type youtubeSearchItemID struct {
	VideoID string `json:"videoId"`
}

type youtubeSnippet struct {
	Title        string `json:"title"`
	ChannelID    string `json:"channelId"`
	ChannelTitle string `json:"channelTitle"`
}

type youtubeVideoResponse struct {
	Items []youtubeVideoItem `json:"items"`
}

type youtubeVideoItem struct {
	Snippet              youtubeSnippet               `json:"snippet"`
	LiveStreamingDetails *youtubeLiveStreamingDetails  `json:"liveStreamingDetails,omitempty"`
}

type youtubeLiveStreamingDetails struct {
	ActualStartTime   string `json:"actualStartTime"`
	ActualEndTime     string `json:"actualEndTime,omitempty"`
	ConcurrentViewers string `json:"concurrentViewers,omitempty"`
}
