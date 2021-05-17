package org

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun-realworld-app/bunapp"
)

type User struct {
	bun.BaseModel `bun:",alias:u"`

	ID           uint64 `json:"-"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Bio          string `json:"bio"`
	Image        string `json:"image"`
	Password     string `bun:"-" json:"password,omitempty"`
	PasswordHash string `json:"-"`
	Following    bool   `bun:"-" json:"following"`

	Token string `bun:"-" json:"token,omitempty"`
}

type FollowUser struct {
	bun.BaseModel `bun:"alias:fu"`

	UserID         uint64
	FollowedUserID uint64
}

type Profile struct {
	bun.BaseModel `bun:"users,alias:u"`

	ID        uint64 `json:"-"`
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `bun:"-" json:"following"`
}

func NewProfile(user *User) *Profile {
	return &Profile{
		Username:  user.Username,
		Bio:       user.Bio,
		Image:     user.Image,
		Following: user.Following,
	}
}

func SelectUser(ctx context.Context, app *bunapp.App, id uint64) (*User, error) {
	user := new(User)
	if err := app.DB().NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func SelectUserByUsername(ctx context.Context, app *bunapp.App, username string) (*User, error) {
	user := new(User)
	if err := app.DB().NewSelect().
		Model(user).
		Where("username = ?", username).
		Scan(ctx); err != nil {
		return nil, err
	}

	return user, nil
}
