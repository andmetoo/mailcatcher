package mailcatcher

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
	"gitlab.com/tozd/go/errors"
)

// Email represents a captured email message.
type Email struct {
	ID      string    `json:"id"`
	From    string    `json:"from"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	Time    time.Time `json:"time"`
	To      []string  `json:"to"`
}

// Logger is a simple logging interface.
type Logger interface {
	Printf(format string, v ...any)
}

// Server is an in-process mail catcher for testing.
type Server struct {
	smtpServer *smtp.Server
	httpServer *http.Server
	logger     Logger
	messages   []Email
	mu         sync.RWMutex
	smtpPort   int
	httpPort   int
}

// New creates a new mail catcher server with custom ports.
func New(smtpPort, httpPort int) *Server {
	s := &Server{
		messages: make([]Email, 0),
		smtpPort: smtpPort,
		httpPort: httpPort,
	}

	// Setup SMTP server
	backend := &backend{server: s}
	s.smtpServer = smtp.NewServer(backend)
	s.smtpServer.Addr = fmt.Sprintf(":%d", smtpPort)
	s.smtpServer.Domain = "localhost"
	s.smtpServer.AllowInsecureAuth = true
	s.smtpServer.MaxLineLength = 16 * 1024 * 1024 // 16MB - allow long lines for HTML emails

	// Setup HTTP API server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/emails", s.handleGetEmails)
	mux.HandleFunc("GET /api/v1/emails/", s.handleGetEmail)
	mux.HandleFunc("DELETE /api/v1/emails", s.handleDeleteEmails)

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

// NewWithDefaults creates a new mail catcher server with default ports.
// SMTP: 1025, HTTP: 8025
func NewWithDefaults() *Server {
	return New(1025, 8025)
}

// Start starts the mail catcher server.
func (s *Server) Start() error {
	lc := &net.ListenConfig{}
	ctx := context.Background()

	// Start SMTP server
	smtpListener, err := lc.Listen(ctx, "tcp", s.smtpServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to start SMTP server: %w", err)
	}

	go func() {
		if serveErr := s.smtpServer.Serve(smtpListener); serveErr != nil && !errors.Is(serveErr, smtp.ErrServerClosed) {
			if s.logger != nil {
				s.logger.Printf("SMTP server error: %v", serveErr)
			}
		}
	}()

	// Start HTTP server
	httpListener, err := lc.Listen(ctx, "tcp", s.httpServer.Addr)
	if err != nil {
		_ = smtpListener.Close()
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	go func() {
		if err := s.httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			if s.logger != nil {
				s.logger.Printf("HTTP server error: %v", err)
			}
		}
	}()

	return nil
}

// Stop stops the mail catcher server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.smtpServer.Close(); err != nil {
		return fmt.Errorf("failed to close SMTP server: %w", err)
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

// Emails returns all captured email messages.
func (s *Server) Emails() []Email {
	s.mu.RLock()
	defer s.mu.RUnlock()

	emails := make([]Email, len(s.messages))
	copy(emails, s.messages)
	return emails
}

// Email returns a specific email by ID.
// Returns nil if email with given ID is not found.
func (s *Server) Email(id string) *Email {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.messages {
		if s.messages[i].ID == id {
			email := s.messages[i]
			return &email
		}
	}
	return nil
}

// Clear removes all captured messages.
func (s *Server) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = make([]Email, 0)
}

// SetLogger sets a custom logger for server errors.
// By default, errors are silently ignored.
func (s *Server) SetLogger(logger Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
}

// addMessage adds a new email to the captured messages.
func (s *Server) addMessage(email Email) {
	s.mu.Lock()
	defer s.mu.Unlock()

	email.ID = fmt.Sprintf("msg-%d", len(s.messages))
	email.Time = time.Now()
	s.messages = append(s.messages, email)
}

// HTTP handlers

func (s *Server) handleGetEmails(w http.ResponseWriter, r *http.Request) {
	emails := s.Emails()

	response := map[string]any{
		"total": len(emails),
		"count": len(emails),
		"items": emails,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleGetEmail(w http.ResponseWriter, r *http.Request) {
	// Extract email ID from path: /api/v1/emails/msg-0
	id := r.URL.Path[len("/api/v1/emails/"):]
	if id == "" {
		http.Error(w, "Email ID is required", http.StatusBadRequest)
		return
	}

	email := s.Email(id)
	if email == nil {
		http.Error(w, "Email not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(email); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) handleDeleteEmails(w http.ResponseWriter, r *http.Request) {
	s.Clear()
	w.WriteHeader(http.StatusNoContent)
}

// SMTP Backend implementation

type backend struct {
	server *Server
}

func (b *backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &session{server: b.server}, nil
}

type session struct {
	server *Server
	from   string
	to     []string
}

func (s *session) AuthPlain(username, password string) error {
	return nil // Accept all auth
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read email data: %w", err)
	}

	// Parse subject from email headers
	subject := parseSubject(body)

	// Store email
	email := Email{
		From:    s.from,
		To:      s.to,
		Subject: subject,
		Body:    string(body),
	}

	s.server.addMessage(email)
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}

// parseSubject extracts the Subject header from email body.
func parseSubject(body []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()

		// Empty line marks end of headers
		if line == "" {
			break
		}

		// Look for Subject header (case-insensitive)
		if strings.HasPrefix(strings.ToLower(line), "subject:") {
			subject := strings.TrimSpace(line[8:]) // Remove "Subject:" prefix
			return subject
		}
	}
	return ""
}

// corsMiddleware adds CORS headers to allow web UI to access the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
