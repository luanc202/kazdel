package handlers

type Handlers struct {
	Home         *HomePageHandler
	Auth         *AuthHandler
	ShortenedUrl *ShortenedUrlHandler
}
