package model

type Link struct {
	FullURL  string `json:"url"`
	Shortcut string `json:"shortcut"`
}

type CreateLinkWithCorrelationIdRequestItem struct {
	FullURL       string `json:"original_url"`
	CorrelationId string `json:"correlation_id"`
}

type CreateLinkWithCorrelationIdResponseItem struct {
	Shortcut      string `json:"short_url"`
	CorrelationId string `json:"correlation_id"`
}
