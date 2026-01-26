package dto

type ShortenedUrlView struct {
	ShortSlug   string `json:"shortSlug"`
	OriginalUrl string `json:"originalUrl"`
}

func NewShortenedUrlView(shortSlug, originalUrl string) *ShortenedUrlView {
	return &ShortenedUrlView{
		ShortSlug:   shortSlug,
		OriginalUrl: originalUrl,
	}
}
