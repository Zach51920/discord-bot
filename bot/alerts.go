package bot

import "log"

func (b *Bot) SendAlert(alert string) {
	log.Println("[ALERT] " + alert)

	b.wg.Add(1)
	defer b.wg.Done()
	b.session.Lock()
	defer b.session.Unlock()
	if _, err := b.session.ChannelMessageSend(b.config.AlertChannelID, alert); err != nil {
		log.Println("[WARNING] failed to write alert:", err)
	}
}
