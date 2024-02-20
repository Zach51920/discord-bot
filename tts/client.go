package tts

import (
	googletts "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"google.golang.org/api/option"
)

const apiKey = "AIzaSyDpnichUaq3woYXzHsJ99ALR4u2JpaR984"

type Client struct {
	client *googletts.Client
}

func NewClient() (*Client, error) {
	ttsClient, err := googletts.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return &Client{}, err
	}
	return &Client{client: ttsClient}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) GetOPUSAudio(text string, voice string) ([]byte, error) {
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en_US",
			Name:         voice,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		//Voice: &texttospeechpb.VoiceSelectionParams{
		//	LanguageCode: "en-US",
		//	SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
		//},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
		},
	}
	resp, err := c.client.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.AudioContent, nil
}
