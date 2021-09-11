package blog

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/org"
)

type Article struct {
	bun.BaseModel `bun:"articles,alias:a"`

	ID          uint64 `json:"-"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`

	Author   *org.Profile `json:"author" bun:"rel:belongs-to"`
	AuthorID uint64       `json:"-"`

	Tags    []ArticleTag `json:"-" bun:"rel:has-many"`
	TagList []string     `json:"tagList" bun:",scanonly,array"`

	Favorited      bool `json:"favorited" bun:",scanonly"`
	FavoritesCount int  `json:"favoritesCount" bun:",scanonly"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ArticleTag struct {
	bun.BaseModel `bun:"alias:t"`

	ArticleID uint64
	Tag       string
}

type FavoriteArticle struct {
	bun.BaseModel `bun:"alias:fa"`

	UserID    uint64
	ArticleID uint64
}

func SelectArticle(ctx context.Context, app *bunapp.App, slug string) (*Article, error) {
	article := new(Article)
	if err := app.DB().NewSelect().
		Model(article).
		Where("slug = ?", slug).
		Scan(ctx); err != nil {
		return nil, err
	}
	return article, nil
}

func selectArticleByFilter(ctx context.Context, app *bunapp.App, f *ArticleFilter) (*Article, error) {
	article := new(Article)
	if err := app.DB().NewSelect().
		Model(article).
		ColumnExpr("?TableColumns").
		Apply(f.query).
		Scan(ctx); err != nil {
		return nil, err
	}

	if article.TagList == nil {
		article.TagList = make([]string, 0)
	}

	return article, nil
}

func createTags(ctx context.Context, app *bunapp.App, article *Article) error {
	if len(article.TagList) == 0 {
		return nil
	}

	tags := make([]ArticleTag, 0, len(article.TagList))
	for _, t := range article.TagList {
		tags = append(tags, ArticleTag{
			ArticleID: article.ID,
			Tag:       t,
		})
	}

	if _, err := app.DB().NewInsert().
		Model(&tags).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
