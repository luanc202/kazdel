package entity

import (
	"time"
)

type UrlVisit struct {
	ID        uint64    `json:"id"`
	UrlId     uint64    `json:"urlId"`
	IpAddress *string   `json:"ipAddress"`
	Referrer  *string   `json:"referrer"`
	UserAgent *string   `json:"userAgent"`
	Browser   *string   `json:"browser"`
	Os        *string   `json:"os"`
	Country   *string   `json:"country"`
	ClickedAt time.Time `json:"clickedAt"`
}

func NewUrlVisit(urlId uint64, ipAddress, referrer, userAgent, browser, os, country *string) *UrlVisit {
	return &UrlVisit{
		UrlId:     urlId,
		IpAddress: ipAddress,
		Referrer:  referrer,
		UserAgent: userAgent,
		Browser:   browser,
		Os:        os,
		Country:   country,
		ClickedAt: time.Now(),
	}
}
