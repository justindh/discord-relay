package listener

import (
	"github.com/bwmarrin/discordgo"

	"github.com/justindh/discord-relay/internal/forwarder"
)

//Listeners all of the incloming updates.
type Listeners struct {
	Sessions []*discordgo.Session
}

// NewListeners bulk creates all listeners and gets them running
func NewListeners(tokens []string, forwarder *forwarder.Forwarder) (Listeners, error) {
	l := Listeners{Sessions: make([]*discordgo.Session, 0)}
	for _, t := range tokens {
		sess, err := newListener(t, forwarder)
		if err != nil {
			return l, err
		}
		l.Sessions = append(l.Sessions, sess)
	}
	return l, nil
}

func newListener(token string, forwarder *forwarder.Forwarder) (*discordgo.Session, error) {
	sess, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	// We want to appear offline
	sess.Identify.Presence.Status = string(discordgo.StatusInvisible)
	// We want to filter to only messages (less clutter)
	sess.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	// Add a message processor
	sess.AddHandler(forwarder.MessageCreate)
	// open the session and start listening for messages
	err = sess.Open()
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// Close attempts to stop and clean up all listeners
func (l *Listeners) Close() {
	for _, s := range l.Sessions {
		s.Close()
	}
}
