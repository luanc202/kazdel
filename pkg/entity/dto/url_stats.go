package dto

type UrlStats struct {
	TotalClicks   int            `json:"totalClicks"`
	BrowserStats  map[string]int `json:"browserStats"`
	OsStats       map[string]int `json:"osStats"`
	CountryStats  map[string]int `json:"countryStats"`
	ReferrerStats map[string]int `json:"referrerStats"`
}

func NewUrlStats() *UrlStats {
	return &UrlStats{
		TotalClicks:   0,
		BrowserStats:  make(map[string]int),
		OsStats:       make(map[string]int),
		CountryStats:  make(map[string]int),
		ReferrerStats: make(map[string]int),
	}
}
