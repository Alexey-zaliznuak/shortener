package model

type Link struct {
	FullURL   string `json:"url"`
	Shortcut  string `json:"shortcut"`
	UserID    string `json:"userID"`
	IsDeleted bool   `json:"isDeleted"`
}

func (l *Link) ToCreateDto() *CreateLinkDto {
	return &CreateLinkDto{
		FullURL:  l.FullURL,
		Shortcut: l.Shortcut,
	}
}

type GetUserLinksRequestItem struct {
	FullURL  string `json:"original_url"`
	Shortcut string `json:"short_url"`
}

type CreateLinkDto struct {
	FullURL  string `json:"url"`
	Shortcut string `json:"shortcut"`
}

func (dto *CreateLinkDto) NewLink(userID string) *Link {
	return &Link{
		FullURL:  dto.FullURL,
		Shortcut: dto.Shortcut,
		UserID:   userID,
	}
}

type CreateLinkWithCorrelationIDRequestItem struct {
	FullURL       string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type CreateLinkWithCorrelationIDResponseItem struct {
	Shortcut      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

type CreateShortURLRequest struct {
	FullURL string `json:"url"`
}
type CreateShortURLResponse struct {
	Result string `json:"result"`
}
