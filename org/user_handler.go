package org

import (
	"errors"
	"net/http"
	"time"

	"github.com/vmihailenco/treemux"
	"golang.org/x/crypto/bcrypt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/httputil"
)

const kb = 10

var errUserNotFound = errors.New("Not registered email or invalid password")

type UserHandler struct {
	app *bunapp.App
}

func NewUserHandler(app *bunapp.App) UserHandler {
	return UserHandler{
		app: app,
	}
}

func (*UserHandler) Current(w http.ResponseWriter, req treemux.Request) error {
	user := UserFromContext(req.Context())
	return treemux.JSON(w, treemux.H{
		"user": user,
	})
}

func (h UserHandler) Create(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	var in struct {
		User *User `json:"user"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := in.User

	var err error
	user.PasswordHash, err = hashPassword(user.Password)
	if err != nil {
		return err
	}

	if _, err := h.app.DB().NewInsert().
		Model(user).
		Exec(ctx); err != nil {
		return err
	}

	if err := setUserToken(h.app, user); err != nil {
		return err
	}

	user.Password = ""
	return treemux.JSON(w, treemux.H{
		"user": user,
	})
}

func (h UserHandler) Login(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	var in struct {
		User *struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}
	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := new(User)
	if err := h.app.DB().NewSelect().
		Model(user).
		Where("email = ?", in.User.Email).
		Scan(ctx); err != nil {
		return err
	}

	if err := comparePasswords(user.PasswordHash, in.User.Password); err != nil {
		return err
	}

	if err := setUserToken(h.app, user); err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"user": user,
	})
}

func (h UserHandler) Update(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	var in struct {
		User *User `json:"user"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.User == nil {
		return errors.New(`JSON field "user" is required`)
	}

	user := in.User

	var err error
	user.PasswordHash, err = hashPassword(user.Password)
	if err != nil {
		return err
	}

	if _, err = h.app.DB().NewUpdate().
		Model(authUser).
		Set("email = ?", user.Email).
		Set("username = ?", user.Username).
		Set("password_hash = ?", user.PasswordHash).
		Set("image = ?", user.Image).
		Set("bio = ?", user.Bio).
		Where("id = ?", authUser.ID).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	user.Password = ""
	return treemux.JSON(w, treemux.H{
		"user": authUser,
	})
}

func (h UserHandler) Profile(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	followingColumn := func(q *bun.SelectQuery) *bun.SelectQuery {
		if authUser, ok := ctx.Value(userCtxKey{}).(*User); ok {
			subq := h.app.DB().NewSelect().
				Model((*FollowUser)(nil)).
				Where("fu.followed_user_id = u.id").
				Where("fu.user_id = ?", authUser.ID)

			q = q.ColumnExpr("EXISTS (?) AS following", subq)
		} else {
			q = q.ColumnExpr("false AS following")
		}

		return q
	}

	user := new(User)
	if err := h.app.DB().NewSelect().
		Model(user).
		ColumnExpr("u.*").
		Apply(followingColumn).
		Where("username = ?", req.Param("username")).
		Scan(ctx); err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"profile": NewProfile(user),
	})
}

func (h UserHandler) Follow(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	user, err := SelectUserByUsername(ctx, h.app, req.Param("username"))
	if err != nil {
		return err
	}

	followUser := &FollowUser{
		UserID:         authUser.ID,
		FollowedUserID: user.ID,
	}
	if _, err := h.app.DB().NewInsert().
		Model(followUser).
		Exec(ctx); err != nil {
		return err
	}

	user.Following = true
	return treemux.JSON(w, treemux.H{
		"profile": NewProfile(user),
	})
}

func (h UserHandler) Unfollow(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	authUser := UserFromContext(ctx)

	user, err := SelectUserByUsername(ctx, h.app, req.Param("username"))
	if err != nil {
		return err
	}

	if _, err := h.app.DB().NewDelete().
		Model((*FollowUser)(nil)).
		Where("user_id = ?", authUser.ID).
		Where("followed_user_id = ?", user.ID).
		Exec(ctx); err != nil {
		return err
	}

	user.Following = false
	return treemux.JSON(w, treemux.H{
		"profile": NewProfile(user),
	})
}

func setUserToken(app *bunapp.App, user *User) error {
	token, err := CreateUserToken(app, user.ID, 24*time.Hour)
	if err != nil {
		return err
	}
	user.Token = token
	return nil
}

func hashPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func comparePasswords(hash, pass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return errUserNotFound
	}
	return nil
}
