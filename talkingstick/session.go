package talkingstick

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"sync"
	"time"
)

type tsSession struct {
	staleTimer   *time.Timer
	ticker       *time.Ticker
	startTime    time.Time
	turnDuration time.Duration

	channelID  string
	isRunning  bool
	shutdownCh chan struct{}
	mu         *sync.Mutex
	quitOnce   sync.Once

	stickholder *tsMember

	embed   *discordgo.Message
	sess    *discordgo.Session
	display *ControlPanel
}

func newTSSession(s *discordgo.Session, channelID string, duration time.Duration, head *tsMember) (*tsSession, error) {
	tss := &tsSession{
		staleTimer:   time.NewTimer(15 * time.Minute),
		ticker:       time.NewTicker(duration),
		startTime:    time.Now(),
		turnDuration: duration,
		channelID:    channelID,
		isRunning:    false,
		shutdownCh:   make(chan struct{}),
		mu:           &sync.Mutex{},
		quitOnce:     sync.Once{},
		stickholder:  head,
		embed:        nil,
		sess:         s,
	}
	panel, err := NewControlPanel(tss)
	tss.display = panel
	return tss, err
}

func (tss *tsSession) Start() {
	slog.Debug("starting talking stick session routine", "channel_id", tss.channelID)

	defer tss.Close()
	for {
		select {
		case <-tss.shutdownCh:
			slog.Debug("received TS quit event")
			return // manually closed session
		case <-tss.staleTimer.C:
			slog.Info("closing TS session due to inactivity", "channel_id", tss.channelID)
			return // session is inactive
		case <-tss.ticker.C:
			slog.Debug("received TS ticker event")
			tss.Pass(nil) // turns up, pass the stick
		}
	}
}

func (tss *tsSession) Pass(target *tsMember) {
	if !tss.Running() {
		slog.Debug("session is paused, don't pass the talking stick")
		return
	}

	tss.mu.Lock()
	if target != nil {
		tss.stickholder = target
	} else {
		tss.stickholder = tss.stickholder.next
	}
	tss.mu.Unlock()

	stickholder := tss.stickholder.data.User
	slog.Debug("passing the talking stick", "stickholder", stickholder.Username)

	// set the priority speaker
	if err := tss.sess.ChannelPermissionSet(tss.channelID, stickholder.ID,
		discordgo.PermissionOverwriteTypeMember, discordgo.PermissionVoicePrioritySpeaker, 0); err != nil {
		slog.Error("failed to set priority speaker", "channel_id", tss.channelID, "user_id", stickholder.ID, "error", err)
	}

	// update the control panel
	tss.display.Refresh()
	tss.resetTicker()
}

func (tss *tsSession) Pause() {
	tss.mu.Lock()
	if tss.isRunning {
		slog.Debug("pausing TS session", "channel_id", tss.channelID)
		tss.isRunning = false
		tss.ticker.Reset(time.Hour)
	}
	tss.mu.Unlock()
	tss.display.Set(DisplayTSPause)
	tss.display.Refresh()
}

func (tss *tsSession) Play() {
	tss.mu.Lock()
	if !tss.isRunning {
		slog.Debug("resuming TS session", "channel_id", tss.channelID)
		tss.isRunning = true
		tss.resetTicker()
	}
	tss.mu.Unlock()
	tss.display.Set(DisplayTSActive)
	tss.display.Refresh()
}

func (tss *tsSession) Quit() {
	tss.quitOnce.Do(func() { close(tss.shutdownCh) })
}

func (tss *tsSession) Close() {
	slog.Debug("closing talking stick session", "channel_id", tss.channelID)

	tss.mu.Lock()
	defer tss.mu.Unlock()

	tss.ticker.Stop()
	tss.staleTimer.Stop()

	// remove priority speaker
	if err := tss.sess.ChannelPermissionSet(tss.channelID, tss.stickholder.data.User.ID,
		discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionVoicePrioritySpeaker); err != nil {
		slog.Error("failed to remove priority speaker", "user_id", tss.stickholder.data.User.Username, "error", err)
	}

	// remove all buttons from the control panel
	tss.display.Set(DisplayTSDecommissioned)
	tss.display.Refresh()
}

func (tss *tsSession) Running() bool {
	tss.mu.Lock()
	defer tss.mu.Unlock()
	return tss.isRunning
}

func (tss *tsSession) resetTicker() {
	tss.ticker.Reset(tss.turnDuration)
}

func (tss *tsSession) resetTimer() {
	tss.staleTimer.Reset(15 * time.Minute)
}
