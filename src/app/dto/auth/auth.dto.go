package auth

import "github.com/golang-jwt/jwt/v4"

type ChulaSSOCredential struct {
	UID         string   `json:"uid"`
	Username    string   `json:"username"`
	Gecos       string   `json:"gecos"`
	Email       string   `json:"email"`
	Disable     bool     `json:"disable"`
	Roles       []string `json:"roles"`
	Firstname   string   `json:"firstname"`
	Lastname    string   `json:"lastname"`
	FirstnameTH string   `json:"firstnameth"`
	LastnameTH  string   `json:"lastnameth"`
	Ouid        string   `json:"ouid"`
}

type TokenPayloadAuth struct {
	jwt.RegisteredClaims
	UserId string `json:"user_id"`
	Role   string `json:"role"`
}
