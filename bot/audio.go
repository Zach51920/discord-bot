package bot

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"io"
	"layeh.com/gopus"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var ErrInvalidChannelType = errors.New("Invalid channel type ")

func (b *Bot) getVoiceConnection(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.VoiceConnection, error) {
	// check for an existing voice connection:
	if voice, ok := b.voiceConnections[i.ChannelID]; ok && voice.Ready {
		return voice, nil
	}

	// if no voice connection exists, create one
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return nil, ErrInvalidChannelType
	}
	var voice *discordgo.VoiceConnection
	if voice, err = s.ChannelVoiceJoin(b.config.GuildID, i.ChannelID, false, false); err != nil {
		return nil, fmt.Errorf("failed to join channel: %w", err)
	}
	// save the voice connection so it can be reused later
	b.voiceConnections[i.ChannelID] = voice
	return voice, nil
}

func writeAudioBytes(voice *discordgo.VoiceConnection, audioData []byte) error {

	filename := "temp/" + uuid.New().String()
	if err := os.WriteFile(filename, audioData, 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	playAudioFromFile(voice, filename, make(chan bool))

	if err := os.Remove(filename); err != nil {
		log.Println("[WARNING] failed to remove temp file: " + err.Error())
	}

	return nil
}

const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

// This is pretty much stolen from DGVoice

func playAudioFromFile(v *discordgo.VoiceConnection, filename string, stop <-chan bool) {

	// Create a shell command "object" to run.
	run := exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegout, err := run.StdoutPipe()
	if err != nil {
		log.Println("[ERROR] failed to run StdoutPipe: " + err.Error())
		return
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	if err = run.Start(); err != nil {
		log.Println("[ERROR] failed to strat StdoutPipe: " + err.Error())
		return
	}
	defer run.Process.Kill()

	// Send "speaking" packet over the voice websocket
	if err = v.Speaking(true); err != nil {
		log.Println("failed to set speaking: " + err.Error())
	}
	defer v.Speaking(false)

	send := make(chan []int16, 2)
	defer close(send)

	closeChan := make(chan bool)
	go func() {
		sendPCM(v, send)
		closeChan <- true
	}()

	for {
		// read data from ffmpeg stdout
		audiobuf := make([]int16, frameSize*channels)
		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return
		}
		if err != nil {
			log.Println("[ERROR] failed to read from ffmpeg stdout: " + err.Error())
			return
		}

		// Send received PCM to the sendPCM channel
		select {
		case send <- audiobuf:
		case <-closeChan:
			return
		}
	}
}

func sendPCM(v *discordgo.VoiceConnection, pcm <-chan []int16) {
	if pcm == nil {
		return
	}

	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		log.Println("[ERROR] failed to create new encoder: " + err.Error())
		return
	}

	for {
		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			return
		}

		// try encoding pcm frame with Opus
		var opus []byte
		if opus, err = opusEncoder.Encode(recv, frameSize, maxBytes); err != nil {
			log.Println("[ERROR] failed to encode frame: " + err.Error())
			return
		}

		if v.Ready == false || v.OpusSend == nil {
			log.Printf("[ERROR] discordgo not ready for opus packets. %+v : %+v", v.Ready, v.OpusSend)
			return
		}
		// send encoded opus data to the sendOpus channel
		v.OpusSend <- opus
	}
}
