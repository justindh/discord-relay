package forwarder

import (
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// Forwarder Wrapper for the discord session
type Forwarder struct {
	*discordgo.Session
	IsWebHook bool
	Channels  map[string]string
}

// NewForwarder takes in a token and returns a Forward Session
func NewForwarder(token string, webhook bool, chans map[string]string) (*Forwarder, error) {
	d, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, err
	}
	d.Identify.Presence.Status = string(discordgo.StatusDoNotDisturb)
	err = d.Open()
	if err != nil {
		return nil, err
	}
	fs := &Forwarder{d, webhook, chans}
	return fs, nil
}

// Send send a message here and we'll figure out how
func (f *Forwarder) Send(text, channelID string) error {
	if f.IsWebHook {
		return f.ToWebhook(text, channelID)
	}
	return f.ToMessage(text, channelID)
}

// ToMessage forwards a message to the specific chan as a message
func (f *Forwarder) ToMessage(text string, channelID string) error {
	_, err := f.ChannelMessageSend(channelID, text)
	if err != nil {
		return fmt.Errorf("Error forwarding to Message %s", err)
	}
	log.Debugf("Set Message with out error?")
	return nil
}

// ToWebhook forwards a message to the specific chan as a webhook
func (f *Forwarder) ToWebhook(text string, channelID string) error {
	ws, err := f.ChannelWebhooks(channelID)
	if err != nil {
		return fmt.Errorf("Error looking up webook %s", err)
	}
	var wh *discordgo.Webhook
	if len(ws) == 0 {
		wh, err = f.WebhookCreate(channelID, "botman", "")
		if err != nil {
			return fmt.Errorf("Error creating webhook %s", err)
		}
	} else {
		wh = ws[0]
	}
	_, err = f.WebhookExecute(
		wh.ID,
		wh.Token,
		true,
		&discordgo.WebhookParams{
			Content:         text,
			Username:        "",
			AvatarURL:       "",
			TTS:             false,
			Embeds:          nil,
			AllowedMentions: nil,
		})

	if err != nil {
		return fmt.Errorf("Error forwarding to webhook %s", err)
	}
	return nil
}

//MessageCreate discord handler for new messages
func (f *Forwarder) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Debugf("new message %s", m.Content)
	var forwardChanID string
	var ok bool
	// Are we listening to this channel?
	if forwardChanID, ok = f.Channels[m.ChannelID]; !ok {
		log.Debugf("didnt find channel: %s", m.ChannelID)
		log.Debugf("all channels: %v", f.Channels)
		return
	}
	// Get sender info
	srcMember, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Debugf("didnt find srcMember: %s", m.Author.Username)
		return
	}
	// Prevent this from relaying @mentions through
	// Todo previously we converted IDs to nice names
	noAtsReg := regexp.MustCompile(`@(\S+)`)
	m.Content = noAtsReg.ReplaceAllString(m.Content, "@ $1")

	// Pull out the links and make them text again
	var links string
	if m.Attachments != nil {
		for _, a := range m.Attachments {
			links += a.URL + " "
		}
	}
	// convert username into a legible format
	var username string
	if srcMember.Nick != "" {
		username = fmt.Sprintf("%s (%s#%s)", srcMember.Nick, srcMember.User.Username, srcMember.User.Discriminator)
	} else {
		username = fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator)
	}

	// Finally send a nicely formated message
	err = f.Send(fmt.Sprintf("**%s**: %s %s", username, m.Content, links), forwardChanID)
	if err != nil {
		log.Errorf("error sending: %s", err)
	}

}
