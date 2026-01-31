package managesieve

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain/sieve"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/ports"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
)

type Session struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	tokenizer *Tokenizer
	repo      ports.ScriptRepository
	userRepo  ports.UserRepository
	logger    *observability.Logger
	tlsConfig *tls.Config

	state State
	user  string // Authenticated user email
	// saslStage removed as unused
}

type State int

const (
	StateUnauth State = iota
	StateAuth
	StateLogout
)

func NewSession(conn net.Conn, repo ports.ScriptRepository, userRepo ports.UserRepository, logger *observability.Logger, tlsConfig *tls.Config) *Session {
	return &Session{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		repo:      repo,
		userRepo:  userRepo,
		logger:    logger,
		tlsConfig: tlsConfig,
		state:     StateUnauth,
	}
}

func (s *Session) Serve() {
	defer s.conn.Close()
	s.tokenizer = NewTokenizer(s.reader)

	s.sendCapabilities()
	s.flush()

	for {
		cmd, err := s.tokenizer.ReadWord()
		if err != nil {
			s.logger.Debug("managesieve connection closed", "reason", err)
			return
		}
		cmd = strings.ToUpper(cmd)

		switch cmd {
		case "LOGOUT":
			s.state = StateLogout
			s.printf("OK \"Logout completed\"\r\n")
		case "CAPABILITY":
			s.sendCapabilities()
		case "NOOP":
			s.printf("OK \"NOOP completed\"\r\n")
		case "AUTHENTICATE":
			s.handleAuthenticate()
		case "PUTSCRIPT":
			s.handlePutScript()
		case "LISTSCRIPTS":
			s.handleListScripts()
		case "GETSCRIPT":
			s.handleGetScript()
		case "DELETESCRIPT":
			s.handleDeleteScript()
		case "SETACTIVE":
			s.handleSetActive()
		case "HAVESPACE":
			s.handleHaveSpace()
		case "RENAMESCRIPT":
			s.handleRenameScript()
		case "CHECKSCRIPT":
			s.handleCheckScript()
		case "STARTTLS":
			s.handleStartTLS()
		default:
			s.printf("NO \"Unknown command: %s\"\r\n", cmd)
			// Consume rest of line to reset parser state?
			// Simple parser might get stuck. In reality we should skip line.
			_, _ = s.reader.ReadString('\n') //nolint:errcheck
		}
		s.flush()

		if s.state == StateLogout {
			break
		}
	}
}

func (s *Session) handleHaveSpace() {
	// HAVESPACE <name> <size>
	// We ignore name usually for quota
	//nolint:errcheck // We ignore args for now
	s.tokenizer.ReadWord() // name
	//nolint:errcheck // We ignore args for now
	s.tokenizer.ReadWord() // size (int)
	// We don't enforce quota yet or assume we have space
	s.printf("OK\r\n")
}

func (s *Session) handleRenameScript() {
	oldName, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing old name\"\r\n")
		return
	}
	newName, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing new name\"\r\n")
		return
	}

	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}

	// Rename logic: Get, Save New, Delete Old
	// Or better, add Rename to repo?
	// For now, manual implementation
	ctx := context.Background()
	script, err := s.repo.Get(ctx, s.user, oldName)
	if err != nil || script == nil {
		s.printf("NO \"Script not found\"\r\n")
		return
	}

	script.Name = newName
	if err := s.repo.Save(ctx, script); err != nil {
		s.printf("NO \"Rename failed\"\r\n")
		return
	}

	if err := s.repo.Delete(ctx, s.user, oldName); err != nil {
		s.logger.Error("Rename delete failed", "error", err)
	}

	s.printf("OK\r\n")
}

func (s *Session) handleCheckScript() {
	// CHECKSCRIPT <content>
	_, err := s.tokenizer.ReadWord() // content
	if err != nil {
		s.printf("NO \"Missing content\"\r\n")
		return
	}
	// We don't check yet, just say OK
	s.printf("OK\r\n")
}

func (s *Session) printf(format string, args ...any) {
	fmt.Fprintf(s.writer, format, args...)
}

func (s *Session) flush() {
	s.writer.Flush()
}

func (s *Session) sendCapabilities() {
	// Ref: RFC 5804
	s.printf("\"IMPLEMENTATION\" \"MailRaven ManageSieve\"\r\n")
	s.printf("\"SASL\" \"PLAIN\"\r\n")
	s.printf("\"SIEVE\" \"fileinto vacation\"\r\n") // Report extensions supported by engine
	if s.tlsConfig != nil {
		s.printf("\"STARTTLS\"\r\n")
	}
	s.printf("OK\r\n")
}

func (s *Session) handleAuthenticate() {
	mech, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing mechanism\"\r\n")
		return
	}
	mech = strings.ToUpper(strings.Trim(mech, "\""))
	if mech != "PLAIN" {
		s.printf("NO \"Unsupported mechanism\"\r\n")
		return
	}

	// Optional initial response
	resp, _ := s.tokenizer.ReadWord() //nolint:errcheck
	if resp == "" {
		// Send challenge
		s.printf("+\r\n")
		s.flush()
		// Read line
		var err error
		resp, err = s.reader.ReadString('\n')
		if err != nil {
			return
		}
		resp = strings.TrimSpace(resp)
	}

	// Decode base64
	// PLAIN: authorization\0authentication\0password
	// Remove quotes if present? RFC says it is a string, handleQuoted handles quotes.
	// But ReadString('\n') might include quotes if sent as "quoted".
	// Usually client sends literal payload for auth? RFC 4616.
	// ManageSieve uses a string argument for initial response.

	// If resp starts with ", unquote it.
	resp = strings.Trim(resp, "\"")

	data, err := base64.StdEncoding.DecodeString(resp)
	if err != nil {
		s.printf("NO \"Invalid base64\"\r\n")
		return
	}

	parts := strings.Split(string(data), "\x00")
	if len(parts) < 3 { // authn\0authz\0pass or \0email\0pass
		s.printf("NO \"Invalid PLAIN format\"\r\n")
		return
	}

	// authz := parts[0]
	email := parts[1]
	if email == "" {
		email = parts[0] // if authz is used as email? Usually parts[1] is authn id.
	}
	password := parts[2]

	// Validate with repo
	user, err := s.userRepo.Authenticate(context.Background(), email, password)
	if err != nil || user == nil {
		s.printf("NO \"Authentication failed\"\r\n")
		return
	}

	s.user = email
	s.state = StateAuth
	s.printf("OK \"Authenticated\"\r\n")
}

func (s *Session) handlePutScript() {
	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}
	name, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing script name\"\r\n")
		return
	}
	content, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing script content\"\r\n")
		return
	}

	script := &sieve.SieveScript{
		UserID:    s.user,
		Name:      name,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Save(context.Background(), script); err != nil {
		s.printf("NO \"Save failed: %v\"\r\n", err)
		return
	}
	s.printf("OK\r\n")
}

func (s *Session) handleListScripts() {
	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}

	scripts, err := s.repo.List(context.Background(), s.user)
	if err != nil {
		s.printf("NO \"List failed\"\r\n")
		return
	}

	for _, script := range scripts {
		if script.IsActive {
			s.printf("SCRIPT \"%s\" ACTIVE\r\n", script.Name)
		} else {
			s.printf("SCRIPT \"%s\"\r\n", script.Name)
		}
	}
	s.printf("OK\r\n")
}

func (s *Session) handleGetScript() {
	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}
	name, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing script name\"\r\n")
		return
	}

	script, err := s.repo.Get(context.Background(), s.user, name)
	if err != nil {
		s.printf("NO \"Script not found\"\r\n")
		return
	}

	s.printf("{%d}\r\n%s\r\nOK\r\n", len(script.Content), script.Content)
}

func (s *Session) deleteScript(name string) {
	if err := s.repo.Delete(context.Background(), s.user, name); err != nil {
		s.printf("NO \"Delete failed\"\r\n")
	} else {
		s.printf("OK\r\n")
	}
}

func (s *Session) handleDeleteScript() {
	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}
	name, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing script name\"\r\n")
		return
	}
	s.deleteScript(name)
}

func (s *Session) handleSetActive() {
	if s.state != StateAuth {
		s.printf("NO \"Not authenticated\"\r\n")
		return
	}
	name, err := s.tokenizer.ReadWord()
	if err != nil {
		s.printf("NO \"Missing script name\"\r\n")
		return
	}

	if err := s.repo.SetActive(context.Background(), s.user, name); err != nil {
		s.printf("NO \"Activate failed\"\r\n")
		return
	}
	s.printf("OK\r\n")
}

func (s *Session) handleStartTLS() {
	if s.tlsConfig == nil {
		s.printf("NO \"TLS not configured\"\r\n")
		return
	}
	s.printf("OK \"Begin TLS negotiation now\"\r\n")
	s.flush()

	// Upgrade connection
	tlsConn := tls.Server(s.conn, s.tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		s.logger.Error("TLS handshake failed", "error", err)
		return
	}

	s.conn = tlsConn
	s.reader = bufio.NewReader(s.conn)
	s.writer = bufio.NewWriter(s.conn)
	s.tokenizer = NewTokenizer(s.reader)
}
