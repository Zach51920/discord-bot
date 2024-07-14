package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Zach51920/discord-bot/config"
	"github.com/kkdai/youtube/v2"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	httpClient *http.Client
	ytClient   youtube.Client
	apiKey     string
}

func New() *Client {
	return &Client{
		httpClient: http.DefaultClient,
		ytClient:   youtube.Client{},
		apiKey:     config.GetString("GOOGLE_API_KEY"),
	}
}

const searchURL = "https://www.googleapis.com/youtube/v3/search"

func (c *Client) Search(query string) (YTSearchResults, error) {
	req := buildURL(searchURL, map[string]any{
		"part": "snippet",
		"q":    url.QueryEscape(query),
		"key":  c.apiKey,
	})
	resp, err := c.httpClient.Get(req)
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

func (c *Client) Download(videoID string) (*File, error) {
	video, err := c.ytClient.GetVideo(videoID)
	if err != nil {
		if errors.Is(err, youtube.ErrInvalidCharactersInVideoID) || errors.Is(err, youtube.ErrVideoIDMinLength) {
			return nil, errors.New("invalid video ID")
		}
		var playbackErr *youtube.ErrPlayabiltyStatus
		if errors.As(err, &playbackErr) {
			return nil, errors.New(playbackErr.Reason)
		}
		return nil, fmt.Errorf("video error: %w", err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	if len(formats) == 0 {
		return nil, fmt.Errorf("no video formats contain audio channels")
	}

	stream, _, err := c.ytClient.GetStream(video, &formats[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	return &File{
		Name:         strings.Join(strings.Split(video.Title, " "), "-") + ".mp4",
		ContentType:  video.Formats[0].MimeType,
		ReaderCloser: stream,
	}, nil
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
