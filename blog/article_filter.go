package blog

import (
	"github.com/go-pg/urlstruct"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/org"
	"github.com/vmihailenco/treemux"
)

type ArticleFilter struct {
	app *bunapp.App

	UserID    uint64
	Author    string
	Tag       string
	Favorited string
	Slug      string
	Feed      bool
	urlstruct.Pager
}

func decodeArticleFilter(app *bunapp.App, req treemux.Request) (*ArticleFilter, error) {
	ctx := req.Context()
	query := req.URL.Query()

	f := &ArticleFilter{
		app: app,

		Tag:       query.Get("tag"),
		Author:    query.Get("author"),
		Favorited: query.Get("favorited"),
		Slug:      req.Param("slug"),
	}

	if user := org.UserFromContext(ctx); user != nil {
		f.UserID = user.ID
	}

	return f, nil
}

func (f *ArticleFilter) query(q *bun.SelectQuery) *bun.SelectQuery {
	q = q.Relation("Author")

	{
		subq := f.app.DB().NewSelect().
			Model((*ArticleTag)(nil)).
			ColumnExpr("array_agg(t.tag)::text[]").
			Where("t.article_id = a.id")

		q = q.ColumnExpr("(?) AS tag_list", subq)
	}

	if f.UserID == 0 {
		q = q.ColumnExpr("false AS favorited")
	} else {
		subq := f.app.DB().NewSelect().
			Model((*FavoriteArticle)(nil)).
			Where("fa.article_id = a.id").
			Where("fa.user_id = ?", f.UserID)

		q = q.ColumnExpr("EXISTS (?) AS favorited", subq)
	}

	q.Apply(authorFollowingColumn(f.app, f.UserID))

	{
		subq := f.app.DB().NewSelect().
			Model((*FavoriteArticle)(nil)).
			ColumnExpr("count(*)").
			Where("fa.article_id = a.id")

		q = q.ColumnExpr("(?) AS favorites_count", subq)
	}

	if f.Author != "" {
		q = q.Where("author.username = ?", f.Author)
	}

	if f.Tag != "" {
		subq := f.app.DB().NewSelect().
			Model((*ArticleTag)(nil)).
			Distinct().
			ColumnExpr("t.article_id").
			Where("t.tag = ?", f.Tag)

		q = q.Where("a.id IN (?)", subq)
	}

	if f.Feed {
		subq := f.app.DB().NewSelect().
			Model((*org.FollowUser)(nil)).
			ColumnExpr("fu.followed_user_id").
			Where("fu.user_id = ?", f.UserID)

		q = q.Where("a.author_id IN (?)", subq)
	} else if f.Slug != "" {
		q = q.Where("a.slug = ?", f.Slug)
	}

	return q
}

func authorFollowingColumn(app *bunapp.App, userID uint64) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if userID == 0 {
			q = q.ColumnExpr("false AS author__following")
		} else {
			subq := app.DB().NewSelect().
				Model((*org.FollowUser)(nil)).
				Where("fu.followed_user_id = author_id").
				Where("fu.user_id = ?", userID)

			q = q.ColumnExpr("EXISTS (?) AS author__following", subq)
		}
		return q
	}
}
