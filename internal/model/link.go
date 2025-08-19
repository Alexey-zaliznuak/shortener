package model

type Link struct {
	FullURL  string `json:"url"`
	Shortcut string `json:"shortcut"`
}

type CreateLinkWithCorrelationIDRequestItem struct {
	FullURL       string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type CreateLinkWithCorrelationIDResponseItem struct {
	Shortcut      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
