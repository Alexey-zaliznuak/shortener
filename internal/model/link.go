package model

// Link represents a URL link model in the URL shortening system.
type Link struct {
	// FullURL contains the full unshortened URL of the link.
	FullURL string `json:"url"`
	// Shortcut contains the short representation of the link.
	Shortcut string `json:"shortcut"`
	// UserID contains the identifier of the user who created the link.
	UserID string `json:"userID"`
	// IsDeleted indicates whether the link is marked as deleted.
	IsDeleted bool `json:"isDeleted"`
}

// ToCreateDto converts Link to CreateLinkDto.
func (l *Link) ToCreateDto() *CreateLinkDto {
	return &CreateLinkDto{
		FullURL:  l.FullURL,
		Shortcut: l.Shortcut,
	}
}

// GetUserLinksRequestItem represents a request item for retrieving user links.
type GetUserLinksRequestItem struct {
	// FullURL contains the original URL.
	FullURL string `json:"original_url"`
	// Shortcut contains the shortened URL.
	Shortcut string `json:"short_url"`
}

// CreateLinkDto represents data for creating a new link.
type CreateLinkDto struct {
	// FullURL contains the full URL to be shortened.
	FullURL string `json:"url"`
	// Shortcut contains the desired short representation of the link.
	Shortcut string `json:"shortcut"`
}

// NewLink creates a new link based on the DTO data and user identifier.
func (dto *CreateLinkDto) NewLink(userID string) *Link {
	return &Link{
		FullURL:  dto.FullURL,
		Shortcut: dto.Shortcut,
		UserID:   userID,
	}
}

// CreateLinkWithCorrelationIDRequestItem represents a request item
// for creating a link with a correlation identifier.
type CreateLinkWithCorrelationIDRequestItem struct {
	// FullURL contains the original URL to be shortened.
	FullURL string `json:"original_url"`
	// CorrelationID contains the identifier for tracking the request.
	CorrelationID string `json:"correlation_id"`
}

// CreateLinkWithCorrelationIDResponseItem represents a response item
// when creating a link with a correlation identifier.
type CreateLinkWithCorrelationIDResponseItem struct {
	// Shortcut contains the created short URL.
	Shortcut string `json:"short_url"`
	// CorrelationID contains the correlation identifier from the request.
	CorrelationID string `json:"correlation_id"`
}

// CreateShortURLRequest represents a request for creating a short URL.
type CreateShortURLRequest struct {
	// FullURL contains the URL to be shortened.
	FullURL string `json:"url"`
}

// CreateShortURLResponse represents a response with the created short URL.
type CreateShortURLResponse struct {
	// Result contains the result - the shortened URL.
	Result string `json:"result"`
}
