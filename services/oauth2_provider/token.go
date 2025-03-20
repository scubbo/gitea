// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package oauth2_provider //nolint

import (
	"fmt"
	"time"

	"code.gitea.io/gitea/modules/timeutil"

	"github.com/golang-jwt/jwt/v5"
)

// Token represents an Oauth grant

// TokenKind represents the type of token for an oauth application
type TokenKind int

const (
	// KindAccessToken is a token with short lifetime to access the api
	KindAccessToken TokenKind = 0
	// KindRefreshToken is token with long lifetime to refresh access tokens obtained by the client
	KindRefreshToken = iota
)

// Token represents a JWT token used to authenticate a client
type Token struct {
	GrantID int64     `json:"gnt"`
	Kind    TokenKind `json:"tt"`
	Counter int64     `json:"cnt,omitempty"`
	jwt.RegisteredClaims
}

// ParseToken parses a signed jwt string
func ParseToken(jwtToken string, signingKey JWTSigningKey) (*Token, error) {
	parsedToken, err := jwt.ParseWithClaims(jwtToken, &Token{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != signingKey.SigningMethod().Alg() {
			return nil, fmt.Errorf("unexpected signing algo: %v", token.Header["alg"])
		}
		return signingKey.VerifyKey(), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	var token *Token
	var ok bool
	if token, ok = parsedToken.Claims.(*Token); !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

// SignToken signs the token with the JWT secret
func (token *Token) SignToken(signingKey JWTSigningKey) (string, error) {
	token.IssuedAt = jwt.NewNumericDate(time.Now())
	return SignToken(token, signingKey)
}

// OIDCToken represents an OpenID Connect id_token
type OIDCToken struct {
	jwt.RegisteredClaims
	Nonce string `json:"nonce,omitempty"`

	// Scope profile
	Name              string             `json:"name,omitempty"`
	PreferredUsername string             `json:"preferred_username,omitempty"`
	Profile           string             `json:"profile,omitempty"`
	Picture           string             `json:"picture,omitempty"`
	Website           string             `json:"website,omitempty"`
	Locale            string             `json:"locale,omitempty"`
	UpdatedAt         timeutil.TimeStamp `json:"updated_at,omitempty"`

	// Scope email
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`

	// Groups are generated by organization and team names
	Groups []string `json:"groups,omitempty"`
}

// SignToken signs an id_token with the (symmetric) client secret key
func (token *OIDCToken) SignToken(signingKey JWTSigningKey) (string, error) {
	token.IssuedAt = jwt.NewNumericDate(time.Now())
	return SignToken(token, signingKey)
}

func SignToken(token jwt.Claims, signingKey JWTSigningKey) (string, error) {
	jwtToken := jwt.NewWithClaims(signingKey.SigningMethod(), token)
	signingKey.PreProcessToken(jwtToken)
	return jwtToken.SignedString(signingKey.SignKey())
}
