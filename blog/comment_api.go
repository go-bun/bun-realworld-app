package blog

import (
	"errors"
	"net/http"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/httputil"
	"github.com/uptrace/bun-realworld-app/org"
	"github.com/uptrace/treemux"
)

type CommentHandler struct {
	app *bunapp.App
}

func NewCommentHandler(app *bunapp.App) CommentHandler {
	return CommentHandler{
		app: app,
	}
}

func (h CommentHandler) List(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	article, err := SelectArticle(ctx, h.app, req.Param("slug"))
	if err != nil {
		return err
	}

	var userID uint64
	if user := org.UserFromContext(ctx); user != nil {
		userID = user.ID
	}

	comments := make([]*Comment, 0)
	if err := h.app.DB().NewSelect().
		Model(&comments).
		ColumnExpr("c.*").
		Relation("Author").
		Apply(authorFollowingColumn(h.app, userID)).
		Where("article_id = ?", article.ID).
		Scan(ctx); err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"comments": comments,
	})
}

func (h CommentHandler) Show(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()

	article, err := SelectArticle(ctx, h.app, req.Param("slug"))
	if err != nil {
		return err
	}

	id, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	var userID uint64
	if user := org.UserFromContext(ctx); user != nil {
		userID = user.ID
	}

	comment := new(Comment)
	if err := h.app.DB().NewSelect().
		Model(comment).
		ColumnExpr("c.*").
		Relation("Author").
		Apply(authorFollowingColumn(h.app, userID)).
		Where("c.id = ?", id).
		Where("article_id = ?", article.ID).
		Scan(ctx); err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"comment": comment,
	})
}

func (h CommentHandler) Create(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	article, err := SelectArticle(ctx, h.app, req.Param("slug"))
	if err != nil {
		return err
	}

	var in struct {
		Comment *Comment `json:"comment"`
	}

	if err := httputil.UnmarshalJSON(w, req, &in, 10<<kb); err != nil {
		return err
	}

	if in.Comment == nil {
		return errors.New(`JSON field "comment" is required`)
	}

	comment := in.Comment

	comment.AuthorID = user.ID
	comment.ArticleID = article.ID
	comment.CreatedAt = h.app.Clock().Now()
	comment.UpdatedAt = h.app.Clock().Now()

	if _, err := h.app.DB().NewInsert().
		Model(comment).
		Exec(ctx); err != nil {
		return err
	}

	comment.Author = org.NewProfile(user)
	return treemux.JSON(w, treemux.H{
		"comment": comment,
	})
}

func (h CommentHandler) Delete(w http.ResponseWriter, req treemux.Request) error {
	ctx := req.Context()
	user := org.UserFromContext(ctx)

	article, err := SelectArticle(ctx, h.app, req.Param("slug"))
	if err != nil {
		return err
	}

	if _, err := h.app.DB().NewDelete().
		Model((*Comment)(nil)).
		Where("author_id = ?", user.ID).
		Where("article_id = ?", article.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
