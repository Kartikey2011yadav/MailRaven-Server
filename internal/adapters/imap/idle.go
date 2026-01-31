package imap

import (
	"fmt"
	"strings"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/notifications"
)

func (s *Session) handleIdle(cmd *Command) {
	if s.state != StateSelected {
		s.send(fmt.Sprintf("%s NO Select mailbox first", cmd.Tag))
		return
	}

	s.send("+ idling")

	// Subscribe to events
	ch := notifications.GlobalHub.Subscribe(s.user.Email)
	defer notifications.GlobalHub.Unsubscribe(s.user.Email, ch)

	// Channel for client input (DONE or error)
	doneCh := make(chan error, 1) // Buffered to avoid leak if we return early

	// Start goroutine to read ONE line (expecting DONE)
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

	// Event loop
	for {
		select {
		case err := <-doneCh:
			if err != nil {
				// Connection died or bad command
				// We can't really recover the session easily if read failed
				// For bad command, we might be able to, but let's assume termination for simplicity on error
				// or just log.
				// If "DONE" received (err == nil), we preserve session.
				if err.Error() == "expected DONE" {
					// We consumed a line that wasn't DONE.
					// In strict mode, maybe terminate. In loose, maybe treat as command?
					// But we are in IDLE handle.
					s.logger.Warn("IDLE received unexpected input", "error", err)
				}
				// If error is IO, loop will break in Serve anyway.
			}
			s.send(fmt.Sprintf("%s OK IDLE terminated", cmd.Tag))
			return

		case evt := <-ch:
			// Handle event
			if evt.Type == notifications.EventNewMessage && s.selectedMailbox != nil && evt.Mailbox == s.selectedMailbox.Name {
				// In real impl, fetch exact counts. For now increment.
				s.selectedMailbox.MessageCount++
				s.send(fmt.Sprintf("* %d EXISTS", s.selectedMailbox.MessageCount))
				s.send(fmt.Sprintf("* %d RECENT", 1))
			}
		}
	}
}
