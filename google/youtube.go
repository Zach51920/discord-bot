package google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

type YTSearchResults struct {
	Items []YTSearchItem `json:"items"`
}

type YTSearchItem struct {
	Kind string `json:"kind"`
	Etag string `json:"etag"`
	ID   struct {
		Kind    string `json:"kind"`
		VideoID string `json:"videoId"`
	} `json:"id"`
	Snippet struct {
		PublishedAt time.Time `json:"publishedAt"`
		ChannelID   string    `json:"channelId"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Thumbnails  struct {
			Default struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"default"`
		} `json:"thumbnails"`
		ChannelTitle         string    `json:"channelTitle"`
		LiveBroadcastContent string    `json:"liveBroadcastContent"`
		PublishTime          time.Time `json:"publishTime"`
	} `json:"snippet"`
}

func (c *Client) SearchYT(query string) (YTSearchResults, error) {
	c.assureClient()

	req := buildURL(searchURL, map[string]any{
		"part": "snippet",
		"q":    url.QueryEscape(query),
		"key":  apiKey,
	})
	resp, err := c.client.Get(req)
	if err != nil {
		return YTSearchResults{}, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return YTSearchResults{}, fmt.Errorf("api error: unexpected status code %d", resp.StatusCode)
	}

	var body []byte
	if body, err = io.ReadAll(resp.Body); err != nil {
		return YTSearchResults{}, fmt.Errorf("response error: %w", err)
	}
	ytResponse := YTSearchResults{}
	if err = json.Unmarshal(body, &ytResponse); err != nil {
		return YTSearchResults{}, fmt.Errorf("response error: %w", err)
	}
	return ytResponse, nil
}

func buildURL(baseUrl string, params map[string]any) string {
	builder := strings.Builder{}
	builder.WriteString(baseUrl)

	prefix := "?"
	for k, v := range params {
		builder.WriteString(fmt.Sprintf("%s%s=%v", prefix, k, v))
		prefix = "&"
	}
	return builder.String()
}
