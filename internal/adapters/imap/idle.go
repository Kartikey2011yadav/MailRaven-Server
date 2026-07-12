package imap

import (
	"context"
	"fmt"
	"strings"
)

func (s *Session) handleIdle(cmd *Command) {
	if s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO Select mailbox first", cmd.Tag))
		return
	}

	s.send("+ idling")

	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	// Subscribe to notifications via the distributed bus
	eventCh, unsubscribe, err := s.notificationBus.Listen(ctx, s.user.Email)
	if err != nil {
		s.send(fmt.Sprintf("%s NO IDLE failed: %v", cmd.Tag, err))
		return
	}
	defer unsubscribe()

	doneCh := make(chan error, 1)

	go func() {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			doneCh <- err
			return
		}
		trim := strings.TrimSpace(strings.ToUpper(line))
		if trim != "DONE" {
			doneCh <- fmt.Errorf("expected DONE, got %s", trim)
			return
		}
		doneCh <- nil
	}()

	for {
		select {
		case err := <-doneCh:
			if err != nil {
				s.logger.Warn("IDLE received unexpected input", "error", err)
			}
			s.send(fmt.Sprintf("%s OK IDLE terminated", cmd.Tag))
			return

		case evt := <-eventCh:
			if evt.EventType == "new_message" && s.selectedMailbox != nil && evt.Mailbox == s.selectedMailbox.Name {
				s.selectedMailbox.MessageCount++
				s.send(fmt.Sprintf("* %d EXISTS", s.selectedMailbox.MessageCount))
				s.send(fmt.Sprintf("* %d RECENT", 1))
			}
		}
	}
}
