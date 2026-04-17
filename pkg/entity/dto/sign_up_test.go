package dto

import (
	"testing"
)

func TestSignUpRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     SignUpRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: SignUpRequest{
				Name:     "Valid Name",
				Username: "valid_user",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: false,
		},
		{
			name: "name too short",
			req: SignUpRequest{
				Name:     "",
				Username: "valid_user",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: true,
			errMsg:  "name must be between 1 and 20 characters",
		},
		{
			name: "username too short",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "usr",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: true,
			errMsg:  "username must be between 4 and 20 characters",
		},
		{
			name: "invalid username characters",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "invalid user!",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: true,
			errMsg:  "username must contain only letters, numbers, dot and underscore",
		},
		{
			name: "invalid email format",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "valid_user",
				Email:    "invalid-email",
				Password: "Password123!",
			},
			wantErr: true,
			errMsg:  "email must be a valid email address",
		},
		{
			name: "password too short",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "valid_user",
				Email:    "test@example.com",
				Password: "Pwd1!",
			},
			wantErr: true,
			errMsg:  "password must be between 8 and 20 characters",
		},
		{
			name: "password missing special char",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "valid_user",
				Email:    "test@example.com",
				Password: "Password123",
			},
			wantErr: true,
			errMsg:  "password must include uppercase, lowercase, numbers and special characters",
		},
		{
			name: "password missing upper char",
			req: SignUpRequest{
				Name:     "Valid",
				Username: "valid_user",
				Email:    "test@example.com",
				Password: "password123!",
			},
			wantErr: true,
			errMsg:  "password must include uppercase, lowercase, numbers and special characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error msg = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}
