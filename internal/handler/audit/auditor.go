package audit

import "time"

// возможно это стоит вынести в слой service?

type ShortUrlAction = string

var (
	ShortUrlActionGet    ShortUrlAction = "follow"
	ShortUrlActionCreate ShortUrlAction = "shorten"
)

type AuditorShortUrlOperation interface {
	Audit(ts int64, action ShortUrlAction, userId string, url string) error
}

type AuditorShortUrlOperationManager struct {
	auditors []AuditorShortUrlOperation
}

func (m *AuditorShortUrlOperationManager) AuditNotify(action ShortUrlAction, userId string, url string) {
	ts := time.Now().Unix()

	for _, auditor := range m.auditors {
		auditor.Audit(ts, action, userId, url)
	}
}

func (m *AuditorShortUrlOperationManager) UseAuditor(newAuditor AuditorShortUrlOperation) {
	m.auditors = append(m.auditors, newAuditor)
}
