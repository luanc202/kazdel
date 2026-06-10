package mocks

import (
	"github.com/stretchr/testify/mock"
)

type EmailService struct {
	mock.Mock
}

func (_m *EmailService) SendEmail(to []string, subject string, body string) error {
	ret := _m.Called(to, subject, body)
	return ret.Error(0)
}

func (_m *EmailService) SendVerificationEmail(to, username, verificationLink string) error {
	ret := _m.Called(to, username, verificationLink)
	return ret.Error(0)
}

func (_m *EmailService) SendPasswordResetEmail(to, username, resetLink string) error {
	ret := _m.Called(to, username, resetLink)
	return ret.Error(0)
}

func (_m *EmailService) SendReportEmail(reportedURL, reason, description, reporterEmail string) error {
	ret := _m.Called(reportedURL, reason, description, reporterEmail)
	return ret.Error(0)
}

func NewEmailService(t interface {
	mock.TestingT
	Cleanup(func())
}) *EmailService {
	m := &EmailService{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}
