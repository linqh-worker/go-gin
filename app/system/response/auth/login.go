package auth

import userResponse "go-gin/app/system/response/user"

type LoginResp struct {
	Token     string                `json:"token"`
	TokenType string                `json:"token_type"`
	ExpiresIn int                   `json:"expires_in"`
	User      userResponse.AuthUser `json:"user"`
}
