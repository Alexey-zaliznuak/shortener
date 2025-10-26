package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type AuditShortUrlOperationHttp struct {
	Url string
}

type AuditShortUrlOperationFile struct {
	FilePath string
	mu       sync.Mutex
}

type AuditPayload struct {
	Ts     int64            `json:"ts"`
	Action ShortUrlAction `json:"action"`
	UserId string         `json:"user_id"`
	Url    string         `json:"url"`
}

func (a *AuditShortUrlOperationHttp) Audit(ts int64, action ShortUrlAction, userId string, url string) error {
	req := AuditPayload{
		Ts:     ts,
		Action: action,
		UserId: userId,
		Url:    url,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(a.Url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *AuditShortUrlOperationFile) Audit(ts int64, action ShortUrlAction, userId string, url string) error {
	req := AuditPayload{
		Ts:     ts,
		Action: action,
		UserId: userId,
		Url:    url,
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
