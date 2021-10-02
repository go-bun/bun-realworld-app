package blog

import (
	"errors"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/httputil"
	"github.com/uptrace/bun-realworld-app/org"
	"github.com/uptrace/bunrouter"

	"github.com/gosimple/slug"
)

const kb = 10

type ArticleHandler struct {
	app *bunapp.App
}

func NewArticleHandler(app *bunapp.App) ArticleHandler {
	return ArticleHandler{
		app: app,
	}
}

func (h ArticleHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(h.app, req)
	if err != nil {
		return err
	}

	articles := make([]*Article, 0)
	if err := h.app.DB().NewSelect().
		Model(&articles).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Limit(f.Pager.GetLimit()).
		Offset(f.Pager.GetOffset()).
		Scan(ctx); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"articles":      articles,
		"articlesCount": len(articles),
	})
}

func (h ArticleHandler) Show(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(h.app, req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, h.app, f)
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"article": article,
	})
}

func (h ArticleHandler) Feed(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	f, err := decodeArticleFilter(h.app, req)
	if err != nil {
		return err
	}
	f.Feed = true

	articles := make([]*Article, 0)
	if err := h.app.DB().NewSelect().
		Model(&articles).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Scan(ctx); err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"articles":      articles,
		"articlesCount": len(articles),
	})
}

func (h ArticleHandler) Create(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	var in struct {
		Article *Article `json:"article"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 100<<kb); err != nil {
		return err
	}

	if in.Article == nil {
		return errors.New(`JSON field "article" is required`)
	}

	article := in.Article

	article.Slug = makeSlug(article.Title)
	article.AuthorID = user.ID
	article.CreatedAt = h.app.Clock().Now()
	article.UpdatedAt = h.app.Clock().Now()

	if _, err := h.app.DB().NewInsert().
		Model(article).
		Exec(ctx); err != nil {
		return err
	}

	if err := createTags(ctx, h.app, article); err != nil {
		return err
	}

	article.Author = org.NewProfile(user)
	return bunrouter.JSON(w, bunrouter.H{
		"article": article,
	})
}

func (h ArticleHandler) Update(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	var in struct {
		Article *Article `json:"article"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 100<<kb); err != nil {
		return err
	}

	if in.Article == nil {
		return errors.New(`JSON field "article" is required`)
	}

	article := in.Article

	if _, err := h.app.DB().NewUpdate().
		Model(article).
		Set("title = ?", article.Title).
		Set("description = ?", article.Description).
		Set("body = ?", article.Body).
		Set("updated_at = ?", h.app.Clock().Now()).
		Where("slug = ?", req.Param("slug")).
		Returning("*").
		Exec(ctx); err != nil {
		return err
	}

	if _, err := h.app.DB().NewDelete().
		Model((*ArticleTag)(nil)).
		Where("article_id = ?", article.ID).
		Exec(ctx); err != nil {
		return err
	}

	if err := createTags(ctx, h.app, article); err != nil {
		return err
	}

	if article.TagList == nil {
		article.TagList = make([]string, 0)
	}

	article.Author = org.NewProfile(user)
	return bunrouter.JSON(w, bunrouter.H{
		"article": article,
	})
}

func (h ArticleHandler) Delete(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	if _, err := h.app.DB().NewDelete().
		Model((*Article)(nil)).
		Where("author_id = ?", user.ID).
		Where("slug = ?", req.Param("slug")).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (h ArticleHandler) Favorite(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f, err := decodeArticleFilter(h.app, req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, h.app, f)
	if err != nil {
		return err
	}

	favoriteArticle := &FavoriteArticle{
		UserID:    user.ID,
		ArticleID: article.ID,
	}
	res, err := h.app.DB().NewInsert().
		Model(favoriteArticle).
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 0 {
		article.Favorited = true
		article.FavoritesCount = article.FavoritesCount + 1
	}

	return bunrouter.JSON(w, bunrouter.H{
		"article": article,
	})
}

func (h ArticleHandler) Unfavorite(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	f, err := decodeArticleFilter(h.app, req)
	if err != nil {
		return err
	}

	article, err := selectArticleByFilter(ctx, h.app, f)
	if err != nil {
		return err
	}

	res, err := h.app.DB().NewDelete().
		Model((*FavoriteArticle)(nil)).
		Where("user_id = ?", user.ID).
		Where("article_id = ?", article.ID).
		Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 0 {
		article.Favorited = false
		article.FavoritesCount = article.FavoritesCount - 1
	}

	return bunrouter.JSON(w, bunrouter.H{
		"article": article,
	})
}

func makeSlug(title string) string {
	return slug.Make(title) + "-" + strconv.Itoa(rand.Int())
}
