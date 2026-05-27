package dto

type ShortenedUrlView struct {
	ShortSlug    string `json:"shortSlug"`
	OriginalUrl  string `json:"originalUrl"`
	ExpiresAt    string `json:"expiresAt"`
	Description  string `json:"description"`
	HasPassword  bool   `json:"hasPassword"`
	RawExpiresAt string `json:"rawExpiresAt"`
	CreatedAt    string `json:"createdAt"`
	Views        string `json:"views"`
}

func NewShortenedUrlView(shortSlug, originalUrl, expiresAt, description string, hasPassword bool, rawExpiresAt, createdAt, views string) *ShortenedUrlView {
	return &ShortenedUrlView{
		ShortSlug:    shortSlug,
		OriginalUrl:  originalUrl,
		ExpiresAt:    expiresAt,
		Description:  description,
		HasPassword:  hasPassword,
		RawExpiresAt: rawExpiresAt,
		CreatedAt:    createdAt,
		Views:        views,
	}
}
