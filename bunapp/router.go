package bunapp

import (
	"net/http"

	"github.com/uptrace/bun-realworld-app/httputil/httperror"
	"github.com/uptrace/treemux"
	"github.com/uptrace/treemux/extra/reqlog"
	"github.com/uptrace/treemux/extra/treemuxgzip"
	"github.com/uptrace/treemux/extra/treemuxotel"
)

func (app *App) initRouter() {
	opts := []treemux.Option{
		treemux.WithMiddleware(treemuxgzip.NewMiddleware()),
		treemux.WithMiddleware(treemuxotel.NewMiddleware()),
	}
	if app.IsDebug() {
		opts = append(opts, treemux.WithMiddleware(reqlog.NewMiddleware()))
	}
	opts = append(opts, treemux.WithMiddleware(errorHandler))

	app.router = treemux.New(opts...)
	app.apiRouter = app.router.NewGroup("/api",
		treemux.WithMiddleware(corsMiddleware),
	)
}

func errorHandler(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}

		httpErr := httperror.From(err)
		if httpErr.Status != 0 {
			w.WriteHeader(httpErr.Status)
		}
		_ = treemux.JSON(w, httpErr)

		return err
	}
}

func corsMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			return next(w, req)
		}

		h := w.Header()

		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")

		// CORS preflight.
		if req.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,HEAD")
			h.Set("Access-Control-Allow-Headers", "authorization,content-type")
			h.Set("Access-Control-Max-Age", "86400")
			return nil
		}

		return next(w, req)
	}
}
