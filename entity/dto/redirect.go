package dto

type Redirect struct {
	LongUrl string `json:"longUrl"`
}

func NewRedirect(longUrl string) *Redirect {
	return &Redirect{
		LongUrl: longUrl,
	}
}
