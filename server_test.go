package mailcatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"testing"
	"time"
)

func TestServerStartStop(t *testing.T) {
	server := NewWithDefaults()

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

func TestSendEmail(t *testing.T) {
	server := New(10025, 10080)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	server.Clear()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Send test email
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	msg := []byte("Subject: Test Email\r\n" +
		"From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"\r\n" +
		"This is a test email body.\r\n")

	err = smtp.SendMail("localhost:10025", nil, from, to, msg)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Wait for email to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify email was captured
	emails := server.Emails()
	if len(emails) != 1 {
		t.Fatalf("Expected 1 email, got %d", len(emails))
	}

	email := emails[0]
	if email.From != from {
		t.Errorf("Expected from=%s, got %s", from, email.From)
	}

	if len(email.To) != 1 || email.To[0] != to[0] {
		t.Errorf("Expected to=%v, got %v", to, email.To)
	}

	if email.Subject != "Test Email" {
		t.Errorf("Expected subject='Test Email', got '%s'", email.Subject)
	}

	if email.Body == "" {
		t.Error("Expected non-empty body")
	}

	if email.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if email.Time.IsZero() {
		t.Error("Expected non-zero time")
	}
}

func TestMultipleEmails(t *testing.T) {
	server := New(10026, 10081)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	server.Clear()
	time.Sleep(100 * time.Millisecond)

	// Send multiple emails
	for i := 0; i < 3; i++ {
		msg := []byte(fmt.Sprintf("Subject: Test %d\r\n\r\nBody %d\r\n", i, i))
		err = smtp.SendMail("localhost:10026", nil, "sender@example.com",
			[]string{"recipient@example.com"}, msg)
		if err != nil {
			t.Fatalf("Failed to send email %d: %v", i, err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	emails := server.Emails()
	if len(emails) != 3 {
		t.Fatalf("Expected 3 emails, got %d", len(emails))
	}

	// Check IDs are unique and sequential
	for i, email := range emails {
		expectedID := fmt.Sprintf("msg-%d", i)
		if email.ID != expectedID {
			t.Errorf("Expected ID=%s, got %s", expectedID, email.ID)
		}
	}
}

func TestEmailMethod(t *testing.T) {
	server := New(10027, 10082)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	server.Clear()
	time.Sleep(100 * time.Millisecond)

	// Send test email
	msg := []byte("Subject: Test\r\n\r\nBody\r\n")
	err = smtp.SendMail("localhost:10027", nil, "sender@example.com",
		[]string{"recipient@example.com"}, msg)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test Email() method
	email := server.Email("msg-0")
	if email == nil {
		t.Fatal("Expected to find email with ID 'msg-0'")
	}

	if email.Subject != "Test" {
		t.Errorf("Expected subject='Test', got '%s'", email.Subject)
	}

	// Test non-existent email
	email = server.Email("msg-999")
	if email != nil {
		t.Error("Expected nil for non-existent email")
	}
}

func TestClearMethod(t *testing.T) {
	server := New(10028, 10083)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Send email
	msg := []byte("Subject: Test\r\n\r\nBody\r\n")
	err = smtp.SendMail("localhost:10028", nil, "sender@example.com",
		[]string{"recipient@example.com"}, msg)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(server.Emails()) == 0 {
		t.Fatal("Expected at least one email before clear")
	}

	server.Clear()

	emails := server.Emails()
	if len(emails) != 0 {
		t.Errorf("Expected 0 emails after clear, got %d", len(emails))
	}
}

func TestHTTPAPI(t *testing.T) {
	server := New(10029, 10084)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	server.Clear()
	time.Sleep(100 * time.Millisecond)

	// Send test email
	msg := []byte("Subject: API Test\r\n\r\nBody content\r\n")
	err = smtp.SendMail("localhost:10029", nil, "sender@example.com",
		[]string{"recipient@example.com"}, msg)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test GET /api/v1/emails
	resp, err := http.Get("http://localhost:10084/api/v1/emails")
	if err != nil {
		t.Fatalf("Failed to GET emails: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	total, ok := result["total"].(float64)
	if !ok || total != 1 {
		t.Errorf("Expected total=1, got %v", result["total"])
	}

	// Test GET /api/v1/emails/{id}
	resp, err = http.Get("http://localhost:10084/api/v1/emails/msg-0")
	if err != nil {
		t.Fatalf("Failed to GET email by ID: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var email Email
	err = json.NewDecoder(resp.Body).Decode(&email)
	if err != nil {
		t.Fatalf("Failed to decode email: %v", err)
	}

	if email.Subject != "API Test" {
		t.Errorf("Expected subject='API Test', got '%s'", email.Subject)
	}

	// Test DELETE /api/v1/emails
	req, err := http.NewRequest(http.MethodDelete, "http://localhost:10084/api/v1/emails", nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to DELETE emails: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}

	// Verify emails are cleared
	if len(server.Emails()) != 0 {
		t.Error("Expected emails to be cleared")
	}
}

func TestHTTPAPINotFound(t *testing.T) {
	server := New(10030, 10085)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Test getting non-existent email
	resp, err := http.Get("http://localhost:10085/api/v1/emails/msg-999")
	if err != nil {
		t.Fatalf("Failed to GET email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestCORSHeaders(t *testing.T) {
	server := New(10031, 10086)
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Test CORS headers
	resp, err := http.Get("http://localhost:10086/api/v1/emails")
	if err != nil {
		t.Fatalf("Failed to GET emails: %v", err)
	}
	defer resp.Body.Close()

	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", corsHeader)
	}

	// Test OPTIONS request (preflight)
	req, err := http.NewRequest(http.MethodOptions, "http://localhost:10086/api/v1/emails", nil)
	if err != nil {
		t.Fatalf("Failed to create OPTIONS request: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", resp.StatusCode)
	}
}
