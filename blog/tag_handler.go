package blog

import (
	"database/sql"
	"net/http"

	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/vmihailenco/treemux"
)

type TagHandler struct {
	app *bunapp.App
}

func NewTagHandler(app *bunapp.App) TagHandler {
	return TagHandler{
		app: app,
	}
}

func (h TagHandler) List(w http.ResponseWriter, req treemux.Request) error {
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

	return treemux.JSON(w, treemux.H{
		"tags": tags,
	})
}
