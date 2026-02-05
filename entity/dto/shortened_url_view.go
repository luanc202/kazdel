package dto

type ShortenedUrlView struct {
	ShortSlug   string `json:"shortSlug"`
	OriginalUrl string `json:"originalUrl"`
	ExpiresAt   string `json:"expiresAt"`
}

func NewShortenedUrlView(shortSlug, originalUrl string, expiresAt string) *ShortenedUrlView {
	return &ShortenedUrlView{
		ShortSlug:   shortSlug,
		OriginalUrl: originalUrl,
		ExpiresAt:   expiresAt,
	}
}
