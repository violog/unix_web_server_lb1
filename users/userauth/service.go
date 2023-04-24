// Copyright (C) 2021 Creditor Corp. Group.
// See LICENSE for copying information.

package userauth

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"

	"todo/pkg/auth"
	"todo/users"
)

const (
	// TokenExpirationTime after passing this time token expires.
	TokenExpirationTime = 24 * time.Hour
)

var (
	// ErrUnauthenticated should be returned when user performs unauthenticated action.
	ErrUnauthenticated = errs.Class("user unauthenticated error")

	// Error is a error class for internal auth errors.
	Error = errs.Class("user auth internal error")
)

// Service is handling all user authentication logic.
//
// architecture: Service
type Service struct {
	users  users.DB
	signer auth.TokenSigner
}

// NewService is a constructor for user auth service.
func NewService(users users.DB, signer auth.TokenSigner) *Service {
	return &Service{
		users:  users,
		signer: signer,
	}
}

// Token authenticates user by credentials and returns auth token.
func (service *Service) Token(ctx context.Context, email string, password string) (token string, err error) {
	user, err := service.users.GetByEmail(ctx, email)
	if err != nil {
		return "", Error.Wrap(err)
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return "", ErrUnauthenticated.Wrap(err)
	}

	claims := auth.Claims{
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: time.Now().UTC().Add(TokenExpirationTime),
	}

	token, err = service.signer.CreateToken(ctx, &claims)
	if err != nil {
		return "", Error.Wrap(err)
	}

	return token, nil
}

// Authorize validates token from context and returns authorized Authorization.
func (service *Service) Authorize(ctx context.Context, tokenS string) (_ auth.Claims, err error) {
	token, err := auth.FromBase64URLString(tokenS)
	if err != nil {
		return auth.Claims{}, Error.Wrap(err)
	}

	claims, err := service.authenticate(token)
	if err != nil {
		return auth.Claims{}, ErrUnauthenticated.Wrap(err)
	}

	err = service.authorize(ctx, claims)
	if err != nil {
		return auth.Claims{}, ErrUnauthenticated.Wrap(err)
	}

	return *claims, nil
}

// authenticate validates token signature and returns authenticated *satelliteauth.Authorization.
func (service *Service) authenticate(token auth.Token) (_ *auth.Claims, err error) {
	signature := token.Signature

	err = service.signer.SignToken(&token)
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare(signature, token.Signature) != 1 {
		return nil, errs.New("incorrect signature")
	}

	claims, err := auth.FromJSON(token.Payload)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// authorize checks claims and returns authorized User.
func (service *Service) authorize(ctx context.Context, claims *auth.Claims) (err error) {
	if !claims.ExpiresAt.IsZero() && claims.ExpiresAt.Before(time.Now()) {
		return ErrUnauthenticated.New("token expiration time has expired")
	}

	_, err = service.users.GetByEmail(ctx, claims.Email)
	if err != nil {
		return errs.New("authorization failed. no user with email: %s", claims.Email)
	}

	return nil
}
