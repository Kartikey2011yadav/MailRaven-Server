package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/smtp/dkim"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/google/uuid"
)

type SendRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type SendHandler struct {
	queueRepo  ports.QueueRepository
	blobStore  ports.BlobStore
	logger     *observability.Logger
	metrics    *observability.Metrics
	dkimSigner *dkim.Signer
	domain     string
}

func NewSendHandler(
	queueRepo ports.QueueRepository,
	blobStore ports.BlobStore,
	logger *observability.Logger,
	metrics *observability.Metrics,
	signingDomain string,
	selector string,
	privateKeyPath string,
) (*SendHandler, error) {
	// If key path is empty, we warn but allow creating handler (sending will fail or be unsigned if we handled that)
	// But requirement implies signing.
	
	if privateKeyPath == "" {
		return nil, fmt.Errorf("DKIM private key path is required")
	}

	pemBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read DKIM key from %s: %w", privateKeyPath, err)
	}
	
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing DKIM key")
	}
	
	var key *rsa.PrivateKey
	
	// Try PKCS1
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = k
	} else {
		// Try PKCS8
		if pk, err2 := x509.ParsePKCS8PrivateKey(block.Bytes); err2 == nil {
			if k, ok := pk.(*rsa.PrivateKey); ok {
				key = k
			} else {
				return nil, fmt.Errorf("key in %s is not an RSA private key", privateKeyPath)
			}
		} else {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	signer := dkim.NewSigner(signingDomain, selector, key)

	return &SendHandler{
		queueRepo:  queueRepo,
		blobStore:  blobStore,
		logger:     logger,
		metrics:    metrics,
		dkimSigner: signer,
		domain:     signingDomain,
	}, nil
}

func (h *SendHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Auth user from context
	email, ok := r.Context().Value(middleware.UserEmailKey).(string)
	if !ok || email == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.To == "" || req.Subject == "" {
		http.Error(w, "To and Subject are required", http.StatusBadRequest)
		return
	}

	// Construct raw MIME message
	messageID := fmt.Sprintf("<%s@%s>", uuid.New().String(), h.domain)
	date := time.Now().UTC().Format(time.RFC1123Z)
	
	// Ensure body has minimal structure
	bodyContent := req.Body
	if !strings.HasSuffix(bodyContent, "\n") {
		bodyContent += "\r\n"
	}
	
	// Create headers
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n",
		email, req.To, req.Subject, date, messageID)
	
	rawMessage := []byte(headers + bodyContent)

	// Sign message
	// Sign standard set of headers
	headersToSign := []string{"From", "To", "Subject", "Date", "Message-ID", "Content-Type", "MIME-Version"}
	
	signatureHeader, err := h.dkimSigner.Sign(rawMessage, headersToSign)
	if err != nil {
		h.logger.Error("failed to sign message", "error", err)
		http.Error(w, "Internal server error during signing", http.StatusInternalServerError)
		return
	}

	// Prepend signature
	signedMessage := []byte(signatureHeader + "\r\n" + string(rawMessage))

	// Store in blob
	blobKey := fmt.Sprintf("outbound/%s.eml", uuid.New().String())
	if err := h.blobStore.Write(context.Background(), blobKey, bytes.NewReader(signedMessage)); err != nil {
		h.logger.Error("failed to write blob", "error", err)
		http.Error(w, "Storage failure", http.StatusInternalServerError)
		return
	}

	// Enqueue
	outMsg := &domain.OutboundMessage{
		ID:          uuid.New().String(),
		Sender:      email,
		Recipient:   req.To,
		BlobKey:     blobKey,
		Status:      domain.QueueStatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		NextRetryAt: time.Now().UTC(),
		RetryCount:  0,
	}

	if err := h.queueRepo.Enqueue(context.Background(), outMsg); err != nil {
		h.logger.Error("failed to enqueue message", "error", err)
		http.Error(w, "Queue failure", http.StatusInternalServerError)
		return
	}
	
	h.metrics.IncrementCounter("messages_outbound_enqueued")

	h.logger.Info("message enqueued", "id", outMsg.ID, "sender", email, "recipient", req.To)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id": outMsg.ID,
		"status": "queued",
		"message_id": messageID,
	})
}
