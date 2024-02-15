package bot

import "log"

// sendAlert sends an alert to the alert channel
func (b *Bot) sendAlert(alert string) {
	log.Println("[ALERT]:", alert)

	b.wg.Add(1)
	defer b.wg.Done()

	b.session.Lock()
	defer b.session.Unlock()
	if _, err := b.session.ChannelMessageSend(b.config.AlertChannelID, alert); err != nil {
		log.Println("[WARNING] failed to write alert:", err)
	}
}
