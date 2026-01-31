package handlers

import (
	"encoding/xml"
	"net/http"

	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type AutodiscoverHandler struct {
	config *config.Config
	logger *observability.Logger
}

func NewAutodiscoverHandler(cfg *config.Config, logger *observability.Logger) *AutodiscoverHandler {
	return &AutodiscoverHandler{
		config: cfg,
		logger: logger,
	}
}

// HandleMozillaAutoconfig serves configuration for Thunderbird
// GET /.well-known/autoconfig/mail/config-v1.1.xml
func (h *AutodiscoverHandler) HandleMozillaAutoconfig(w http.ResponseWriter, r *http.Request) {
	// Default to configured domain if no email (or extract domain from email)
	confDomain := h.config.Domain
	username := "%EMAILADDRESS%"

	resp := dto.ClientConfig{
		Version: "1.1",
		EmailProvider: dto.EmailProvider{
			ID:          confDomain,
			Domain:      confDomain,
			DisplayName: "MailRaven",
			IncomingServer: dto.IncomingServer{
				Type:           "imap",
				Hostname:       h.config.SMTP.Hostname,
				Port:           h.config.IMAP.Port,
				SocketType:     "STARTTLS",
				Authentication: "password-cleartext",
				Username:       username,
			},
			OutgoingServer: dto.OutgoingServer{
				Type:           "smtp",
				Hostname:       h.config.SMTP.Hostname,
				Port:           h.config.SMTP.Port,
				SocketType:     "STARTTLS",
				Authentication: "password-cleartext",
				Username:       username,
			},
		},
	}

	if h.config.IMAP.Port == 993 {
		resp.EmailProvider.IncomingServer.SocketType = "SSL"
	}
	if h.config.SMTP.Port == 465 {
		resp.EmailProvider.OutgoingServer.SocketType = "SSL"
	}

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck // Best effort
	w.Write([]byte(xml.Header))
	if err := xml.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode autoconfig response", "error", err)
	}
}

// HandleMicrosoftAutodiscover serves configuration for Outlook
// POST /autodiscover/autodiscover.xml
func (h *AutodiscoverHandler) HandleMicrosoftAutodiscover(w http.ResponseWriter, r *http.Request) {
	var req dto.AutodiscoverRequest
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode autodiscover request", "error", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	imapProto := dto.Protocol{
		Type:         "IMAP",
		Server:       h.config.SMTP.Hostname,
		Port:         h.config.IMAP.Port,
		SSL:          "on",
		AuthRequired: "on",
	}

	smtpProto := dto.Protocol{
		Type:         "SMTP",
		Server:       h.config.SMTP.Hostname,
		Port:         h.config.SMTP.Port,
		SSL:          "on",
		AuthRequired: "on",
	}

	resp := dto.AutodiscoverResponse{
		Response: dto.Response{
			Account: dto.Account{
				AccountType: "email",
				Action:      "settings",
				Protocol:    []dto.Protocol{imapProto, smtpProto},
			},
		},
	}

	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck // Best effort
	w.Write([]byte(xml.Header))
	if err := xml.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode autodiscover response", "error", err)
	}
}
