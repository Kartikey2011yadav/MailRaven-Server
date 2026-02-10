package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	httpAdapter "github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/dto"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/http/middleware"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/disk"
	"github.com/Kartikey2011yadav/mailraven-server/internal/adapters/storage/sqlite"
	"github.com/Kartikey2011yadav/mailraven-server/internal/config"
	"github.com/Kartikey2011yadav/mailraven-server/internal/core/domain"
	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// testEnvironment holds test infrastructure
type testEnvironment struct {
	server    *httptest.Server
	tempDir   string
	emailRepo *sqlite.EmailRepository
	userRepo  *sqlite.UserRepository
	queueRepo *sqlite.QueueRepository
	blobStore *disk.BlobStore
	messages  []*domain.Message
	conn      *sqlite.Connection
}

// setupTestEnvironment creates test database and server
func setupTestEnvironment(t *testing.T) *testEnvironment {
	// Create temp directory in current dir (not system temp - avoids Windows fsync issues)
	tempDir := filepath.Join(".", "testdata", fmt.Sprintf("test-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Setup test config
	dkimKeyPath := filepath.Join(tempDir, "dkim.pem")
	generateTestDKIMKey(t, dkimKeyPath)

	cfg := &config.Config{
		Domain: "test.example.com",
		API: config.APIConfig{
			Host:      "127.0.0.1",
			Port:      8443,
			JWTSecret: "test-secret-key-for-testing-only",
		},
		Storage: config.StorageConfig{
			DBPath:   filepath.Join(tempDir, "test.db"),
			BlobPath: filepath.Join(tempDir, "blobs"),
		},
		DKIM: config.DKIMConfig{
			Selector:       "default",
			PrivateKeyPath: dkimKeyPath,
		},
	}

	// Create blob storage directory
	if err := os.MkdirAll(cfg.Storage.BlobPath, 0750); err != nil {
		t.Fatalf("Failed to create blob dir: %v", err)
	}

	// Initialize logger and metrics
	logger := observability.NewLogger("error", "json") // Quiet during tests
	metrics := observability.NewMetrics()

	// Initialize database
	conn, err := sqlite.NewConnection(cfg.Storage.DBPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Run migrations
	migrationsDir := "../internal/adapters/storage/sqlite/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("Failed to read migrations dir: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			migrationPath := filepath.Join(migrationsDir, file.Name())
			if err := conn.RunMigrations(migrationPath); err != nil {
				t.Logf("Migration warning (%s): %v", file.Name(), err)
			}
		}
	}
	migrationPath004 := "../internal/adapters/storage/sqlite/migrations/004_add_imap_fields.sql"
	if err := conn.RunMigrations(migrationPath004); err != nil {
		t.Logf("Migration warning (004): %v", err)
	}
	migrationPath005 := "../internal/adapters/storage/sqlite/migrations/005_add_tls_reports.sql"
	if err := conn.RunMigrations(migrationPath005); err != nil {
		t.Logf("Migration warning (005): %v", err)
	}
	migrationPath006 := "../internal/adapters/storage/sqlite/migrations/006_add_greylist.sql"
	if err := conn.RunMigrations(migrationPath006); err != nil {
		t.Logf("Migration warning (006): %v", err)
	}
	migrationPath007 := "../internal/adapters/storage/sqlite/migrations/007_add_bayes.sql"
	if err := conn.RunMigrations(migrationPath007); err != nil {
		t.Logf("Migration warning (007): %v", err)
	}
	migrationPath008 := "../internal/adapters/storage/sqlite/migrations/008_remove_unique_message_id.sql"
	if err := conn.RunMigrations(migrationPath008); err != nil {
		t.Logf("Migration warning (008): %v", err)
	}
	migrationPath010 := "../internal/adapters/storage/sqlite/migrations/010_sieve.sql"
	if err := conn.RunMigrations(migrationPath010); err != nil {
		t.Logf("Migration warning (010): %v", err)
	}
	migrationPath011 := "../internal/adapters/storage/sqlite/migrations/011_add_starred_column.sql"
	if err := conn.RunMigrations(migrationPath011); err != nil {
		t.Logf("Migration warning (011): %v", err)
	}

	// Initialize repositories
	emailRepo := sqlite.NewEmailRepository(conn.DB)
	userRepo := sqlite.NewUserRepository(conn.DB)
	queueRepo := sqlite.NewQueueRepository(conn.DB)
	domainRepo := sqlite.NewDomainRepository(conn.DB)
	searchIdx := sqlite.NewSearchRepository(conn.DB)
	blobStore, err := disk.NewBlobStore(cfg.Storage.BlobPath)
	if err != nil {
		t.Fatalf("Failed to create blob store: %v", err)
	}

	// Create test user
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	testUser := &domain.User{
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
	}
	if err := userRepo.Create(context.Background(), testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test messages
	messages := []*domain.Message{
		{
			ID:          "msg-1",
			MessageID:   "<test1@example.com>",
			Sender:      "sender1@example.com",
			Recipient:   "test@example.com",
			Subject:     "Test Message 1",
			Snippet:     "This is test message 1",
			ReadState:   false,
			ReceivedAt:  time.Now().Add(-2 * time.Hour),
			SPFResult:   "pass",
			DKIMResult:  "pass",
			DMARCResult: "pass",
		},
		{
			ID:          "msg-2",
			MessageID:   "<test2@example.com>",
			Sender:      "sender2@example.com",
			Recipient:   "test@example.com",
			Subject:     "Test Message 2",
			Snippet:     "This is test message 2",
			ReadState:   false,
			ReceivedAt:  time.Now().Add(-1 * time.Hour),
			SPFResult:   "pass",
			DKIMResult:  "pass",
			DMARCResult: "pass",
		},
		{
			ID:          "msg-3",
			MessageID:   "<test3@example.com>",
			Sender:      "sender3@example.com",
			Recipient:   "test@example.com",
			Subject:     "Test Message 3",
			Snippet:     "This is test message 3",
			ReadState:   false,
			ReceivedAt:  time.Now().Add(-30 * time.Minute),
			SPFResult:   "pass",
			DKIMResult:  "pass",
			DMARCResult: "pass",
		},
	}

	for i, msg := range messages {
		// Create test blob first to get the path
		testBody := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\nTest body content",
			msg.Sender, msg.Recipient, msg.Subject)
		path, err := blobStore.Write(context.Background(), msg.ID, []byte(testBody))
		if err != nil {
			t.Fatalf("Failed to write test blob: %v", err)
		}

		// Update message with blob path before saving to DB
		messages[i].BodyPath = path

		// Save message to database
		if err := emailRepo.Save(context.Background(), msg); err != nil {
			t.Fatalf("Failed to save test message: %v", err)
		}
	}

	// Create TLSRpt Repo for tests
	tlsRptRepo := sqlite.NewTLSRptRepository(conn.DB)

	// Create HTTP server
	httpServer := httpAdapter.NewServer(cfg, emailRepo, userRepo, queueRepo, domainRepo, blobStore, searchIdx, nil, nil, tlsRptRepo, nil, nil, &NoOpSpamFilter{}, logger, metrics)
	testServer := httptest.NewServer(httpServer.Router())

	return &testEnvironment{
		server:    testServer,
		tempDir:   tempDir,
		conn:      conn,
		emailRepo: emailRepo,
		userRepo:  userRepo,
		queueRepo: queueRepo,
		blobStore: blobStore,
		messages:  messages,
	}
}

// cleanup removes test data
func (e *testEnvironment) cleanup() {
	e.server.Close()
	os.RemoveAll(e.tempDir)
}

// authenticateUser logs in and returns JWT token
func (e *testEnvironment) authenticateUser(t *testing.T, email, password string) string {
	loginReq := dto.LoginRequest{
		Email:    email,
		Password: password,
	}
	body := e.encodeJSON(t, loginReq)

	req, err := http.NewRequest("POST", e.server.URL+"/api/v1/auth/login", body)
	if err != nil {
		t.Fatalf("Failed to create login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status %d", resp.StatusCode)
	}

	var loginResp dto.LoginResponse
	e.decodeJSON(t, resp.Body, &loginResp)

	return loginResp.Token
}

// newRequest creates an HTTP request with optional auth token
func (e *testEnvironment) newRequest(t *testing.T, method, path string, body io.Reader, token string) *http.Request {
	req, err := http.NewRequest(method, e.server.URL+path, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

// doRequest executes an HTTP request
func (e *testEnvironment) doRequest(t *testing.T, req *http.Request) *http.Response {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// encodeJSON marshals object to JSON reader
func (e *testEnvironment) encodeJSON(t *testing.T, v interface{}) io.Reader {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	return bytes.NewReader(data)
}

// decodeJSON unmarshals JSON from reader
func (e *testEnvironment) decodeJSON(t *testing.T, r io.Reader, v interface{}) {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}
}

// boolPtr returns pointer to bool value
func boolPtr(b bool) *bool {
	return &b
}

// generateExpiredToken creates a JWT token with specified expiration time
func generateExpiredToken(t *testing.T, email string, expiresAt time.Time) string {
	// JWT secret from test config (matches what quickstart generates)
	jwtSecret := "test-secret-key-for-testing-only"

	claims := &middleware.Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-8 * 24 * time.Hour)), // Issued 8 days ago
			Issuer:    "mailraven",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	return tokenString
}

func generateTestDKIMKey(t *testing.T, path string) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	pemFile, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create DKIM key file: %v", err)
	}
	defer pemFile.Close()

	if err := pem.Encode(pemFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		t.Fatalf("Failed to write DKIM key: %v", err)
	}
}

// NoOpSpamFilter for tests
type NoOpSpamFilter struct{}

func (n *NoOpSpamFilter) CheckConnection(ctx context.Context, ip string) error { return nil }
func (n *NoOpSpamFilter) CheckContent(ctx context.Context, content io.Reader, headers map[string]string) (*domain.SpamCheckResult, error) {
	return &domain.SpamCheckResult{Action: domain.SpamActionPass, Score: 0.0}, nil
}
func (n *NoOpSpamFilter) CheckRecipient(ctx context.Context, ip, sender, recipient string) error {
	return nil
}
func (n *NoOpSpamFilter) TrainSpam(ctx context.Context, content io.Reader) error { return nil }
func (n *NoOpSpamFilter) TrainHam(ctx context.Context, content io.Reader) error  { return nil }
