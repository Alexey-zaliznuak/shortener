package audit

import "time"

// возможно это стоит вынести в слой service?

type ShortURLAction = string

var (
	ShortURLActionGet    ShortURLAction = "follow"
	ShortURLActionCreate ShortURLAction = "shorten"
)

type AuditorShortURLOperation interface {
	Audit(ts int64, action ShortURLAction, userID string, url string) error
}

type AuditorShortURLOperationManager struct {
	auditors []AuditorShortURLOperation
}

func (m *AuditorShortURLOperationManager) AuditNotify(action ShortURLAction, userID string, url string) {
	ts := time.Now().Unix()

	for _, auditor := range m.auditors {
		auditor.Audit(ts, action, userID, url)
	}
}

func (m *AuditorShortURLOperationManager) UseAuditor(newAuditor AuditorShortURLOperation) {
	m.auditors = append(m.auditors, newAuditor)
}

func NewAuditorShortURLOperationManager() *AuditorShortURLOperationManager {
	return &AuditorShortURLOperationManager{
		auditors: make([]AuditorShortURLOperation, 0),
	}
}
