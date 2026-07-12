package sieve

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"git.sr.ht/~emersion/go-sieve"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
)

const (
	maxScriptSize      = 100 * 1024 // 100KB
	sieveExecTimeout   = 5 * time.Second
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

// Execute runs the Sieve interpreter with resource limits.
func (e *SieveEngine) Execute(ctx context.Context, userID string, rawMsg []byte) ([]string, error) {
	script, err := e.scriptRepo.GetActive(ctx, userID)
	if err != nil {
		log.Printf("failed to get active script for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}
	if script == nil {
		return []string{"INBOX"}, nil
	}

	if len(script.Content) > maxScriptSize {
		log.Printf("sieve script too large for user %s: %d bytes", userID, len(script.Content))
		return []string{"INBOX"}, fmt.Errorf("script exceeds maximum size of %d bytes", maxScriptSize)
	}

	cmds, err := sieve.Parse(strings.NewReader(script.Content))
	if err != nil {
		log.Printf("failed to parse script %s: %v", script.Name, err)
		return []string{"INBOX"}, nil
	}

	msg, err := mail.ReadMessage(bytes.NewReader(rawMsg))
	if err != nil {
		log.Printf("failed to parse email header for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}

	execCtx, cancel := context.WithTimeout(ctx, sieveExecTimeout)
	defer cancel()

	interp := NewInterpreter(execCtx, msg, e.vacationManager, userID)
	targets, err := interp.Run(cmds)
	if err != nil {
		log.Printf("sieve runtime error for user %s: %v", userID, err)
		return []string{"INBOX"}, nil
	}

	for _, folder := range targets {
		if folder == "INBOX" {
			continue
		}
		if err := e.mailboxRepo.CreateMailbox(ctx, userID, folder); err != nil {
			log.Printf("sieve: failed to create mailbox %s: %v", folder, err)
		}
	}

	return targets, nil
}
