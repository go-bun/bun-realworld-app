package org

import (
	"context"

	"github.com/uptrace/bun-realworld-app/app"
)

func init() {
	app.OnStart("org.initRoutes", func(ctx context.Context, app *app.App) error {
		middleware := NewMiddleware(app)
		userHandler := NewUserHandler(app)

		g := app.APIRouter().WithMiddleware(middleware.User)

		g.POST("/users", userHandler.Create)
		g.POST("/users/login", userHandler.Login)
		g.GET("/profiles/:username", userHandler.Profile)

		g = g.WithMiddleware(middleware.MustUser)

		g.GET("/user/", userHandler.Current)
		g.PUT("/user/", userHandler.Update)

		g.POST("/profiles/:username/follow", userHandler.Follow)
		g.DELETE("/profiles/:username/follow", userHandler.Unfollow)

		return nil
	})
}
