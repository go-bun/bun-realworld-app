package org

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bunrouter"
)

type (
	userCtxKey    struct{}
	userErrCtxKey struct{}
)

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userCtxKey{}).(*User)
	return user
}

func authToken(req bunrouter.Request) string {
	const prefix = "Token "
	v := req.Header.Get("Authorization")
	v = strings.TrimPrefix(v, prefix)
	return v
}

type Middleware struct {
	app *bunapp.App
}

func NewMiddleware(app *bunapp.App) Middleware {
	return Middleware{
		app: app,
	}
}

func (m Middleware) User(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()

		token := authToken(req)
		userID, err := decodeUserToken(m.app, token)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		user, err := SelectUser(ctx, m.app, userID)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		user.Token, err = CreateUserToken(m.app, user.ID, 24*time.Hour)
		if err != nil {
			ctx = context.WithValue(ctx, userErrCtxKey{}, err)
			return next(w, req.WithContext(ctx))
		}

		ctx = context.WithValue(ctx, userCtxKey{}, user)
		return next(w, req.WithContext(ctx))
	}
}

func (m Middleware) MustUser(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		if err, ok := req.Context().Value(userErrCtxKey{}).(error); ok {
			return err
		}
		return next(w, req)
	}
}
