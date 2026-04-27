package dto

type ShortenedUrlView struct {
	ShortSlug    string `json:"shortSlug"`
	OriginalUrl  string `json:"originalUrl"`
	ExpiresAt    string `json:"expiresAt"`
	Description  string `json:"description"`
	HasPassword  bool   `json:"hasPassword"`
	RawExpiresAt string `json:"rawExpiresAt"`
}

func NewShortenedUrlView(shortSlug, originalUrl, expiresAt, description string, hasPassword bool, rawExpiresAt string) *ShortenedUrlView {
	return &ShortenedUrlView{
		ShortSlug:    shortSlug,
		OriginalUrl:  originalUrl,
		ExpiresAt:    expiresAt,
		Description:  description,
		HasPassword:  hasPassword,
		RawExpiresAt: rawExpiresAt,
	}
}
