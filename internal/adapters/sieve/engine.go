package sieve

import (
	"bytes"
	"context"
	"log"
	"net/mail"
	"strings"

	"git.sr.ht/~emersion/go-sieve"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

type SieveEngine struct {
	scriptRepo      ports.ScriptRepository
	mailboxRepo     ports.EmailRepository // For checking/creating folders
	vacationManager *VacationManager
}

func NewSieveEngine(
	scriptRepo ports.ScriptRepository,
	mailboxRepo ports.EmailRepository,
	vacationRepo ports.VacationRepository,
	queueRepo ports.QueueRepository,
	blobStore ports.BlobStore,
) *SieveEngine {
	return &SieveEngine{
		scriptRepo:      scriptRepo,
		mailboxRepo:     mailboxRepo,
		vacationManager: NewVacationManager(vacationRepo, queueRepo, blobStore),
	}
}

// Execute runs the Sieve interpreter.
func (e *SieveEngine) Execute(ctx context.Context, userID string, rawMsg []byte) ([]string, error) {
	// 1. Get active script
	script, err := e.scriptRepo.GetActive(ctx, userID)
	if err != nil {
		log.Printf("failed to get active script for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}
	if script == nil {
		return []string{"INBOX"}, nil
	}

	// 2. Parse script used emersion/go-sieve parser
	cmds, err := sieve.Parse(strings.NewReader(script.Content))
	if err != nil {
		log.Printf("failed to parse script %s: %v", script.Name, err)
		return []string{"INBOX"}, nil // Fail open
	}

	// 3. Parse Message Header
	msg, err := mail.ReadMessage(bytes.NewReader(rawMsg))
	if err != nil {
		log.Printf("failed to parse email header for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}

	// 4. Run Interpreter
	interp := NewInterpreter(ctx, msg, e.vacationManager, userID)
	targets, err := interp.Run(cmds)
	if err != nil {
		log.Printf("sieve runtime error for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}

	// 5. Ensure mailboxes exist (Side Effect)
	for _, folder := range targets {
		if folder == "INBOX" {
			continue
		}
		if err := e.mailboxRepo.CreateMailbox(ctx, userID, folder); err != nil {
			// Log but proceed
			log.Printf("sieve: failed to create mailbox %s: %v", folder, err)
		}
	}

	return targets, nil
}
