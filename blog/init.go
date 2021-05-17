package blog

import (
	"context"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/org"
)

func init() {
	bunapp.OnStart("blog.initRoutes", func(ctx context.Context, app *bunapp.App) error {
		middleware := org.NewMiddleware(app)
		tagHandler := NewTagHandler(app)
		articleHandler := NewArticleHandler(app)
		commentHandler := NewCommentHandler(app)

		g := app.APIRouter().WithMiddleware(middleware.User)

		g.GET("/tags/", tagHandler.List)

		g.GET("/articles", articleHandler.List)
		g.GET("/articles/feed", articleHandler.Feed)
		g.GET("/articles/:slug", articleHandler.Show)

		g.GET("/articles/:slug/comments", commentHandler.List)
		g.GET("/articles/:slug/comments/:id", commentHandler.Show)

		g = g.WithMiddleware(middleware.MustUser)

		g.POST("/articles", articleHandler.Create)
		g.PUT("/articles/:slug", articleHandler.Update)
		g.DELETE("/articles/:slug", articleHandler.Delete)

		g.POST("/articles/:slug/favorite", articleHandler.Favorite)
		g.DELETE("/articles/:slug/favorite", articleHandler.Unfavorite)

		g.POST("/articles/:slug/comments", commentHandler.Create)
		g.DELETE("/articles/:slug/comments/:id", commentHandler.Delete)

		return nil
	})
}
