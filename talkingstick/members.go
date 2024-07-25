package talkingstick

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"math/rand"
)

type tsMember struct {
	data *discordgo.Member
	next *tsMember
}

func loadVoiceMembers(s *discordgo.Session, guildID, channelID string) []*discordgo.Member {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		slog.Error("failed to access guild state", "error", err)
		return nil
	}
	var members []*discordgo.Member
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID {
			member, err := s.GuildMember(guildID, vs.UserID)
			if err != nil {
				slog.Error("failed to fetch tsMember", "user_id", vs.UserID, "error", err)
				continue
			}
			members = append(members, member)
		}
	}
	return members
}

func shuffleDGMembers(members []*discordgo.Member) {
	for i := len(members) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		members[i], members[j] = members[j], members[i]
	}
}

func newMemberList(members []*discordgo.Member) *tsMember {
	if len(members) == 0 {
		return nil
	}

	head := &tsMember{data: members[0]}
	current := head
	for i := 1; i < len(members); i++ {
		newNode := &tsMember{data: members[i]}
		current.next = newNode
		current = newNode
	}
	current.next = head //  make the list circular
	return head
}
