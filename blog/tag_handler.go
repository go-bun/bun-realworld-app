package blog

import (
	"database/sql"
	"net/http"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bunrouter"
)

type TagHandler struct {
	app *bunapp.App
}

func NewTagHandler(app *bunapp.App) TagHandler {
	return TagHandler{
		app: app,
	}
}

func (h TagHandler) List(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	tags := make([]string, 0)
	if err := h.app.DB().NewSelect().
		Model((*ArticleTag)(nil)).
		ColumnExpr("tag").
		GroupExpr("tag").
		OrderExpr("count(tag) DESC").
		Scan(ctx, &tags); err != nil && err != sql.ErrNoRows {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"tags": tags,
	})
}
