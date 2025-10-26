package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type AuditShortURLOperationHTTP struct {
	URL string
}

type AuditShortURLOperationFile struct {
	FilePath string
	mu       sync.Mutex
}

type AuditPayload struct {
	TS     int64          `json:"ts"`
	Action ShortURLAction `json:"action"`
	UserID string         `json:"user_id"`
	URL    string         `json:"url"`
}

func (a *AuditShortURLOperationHTTP) Audit(ts int64, action ShortURLAction, userID string, url string) error {
	req := AuditPayload{
		TS:     ts,
		Action: action,
		UserID: userID,
		URL:    url,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(a.URL, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *AuditShortURLOperationFile) Audit(ts int64, action ShortURLAction, userID string, url string) error {
	req := AuditPayload{
		TS:     ts,
		Action: action,
		UserID: userID,
		URL:    url,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.OpenFile(a.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data = append(data, '\n')

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
