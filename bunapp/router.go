package bunapp

import (
	"net/http"

	"github.com/uptrace/bun-realworld-app/httputil/httperror"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/bunroutergzip"
	"github.com/uptrace/bunrouter/extra/bunrouterotel"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func (app *App) initRouter() {
	opts := []bunrouter.Option{
		bunrouter.WithMiddleware(bunroutergzip.NewMiddleware()),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware()),
	}
	if app.IsDebug() {
		opts = append(opts, bunrouter.WithMiddleware(reqlog.NewMiddleware()))
	}
	opts = append(opts, bunrouter.WithMiddleware(errorHandler))

	app.router = bunrouter.New(opts...)
	app.apiRouter = app.router.NewGroup("/api",
		bunrouter.WithMiddleware(corsMiddleware),
	)
}

func errorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err == nil {
			return nil
		}

		httpErr := httperror.From(err)
		if httpErr.Status != 0 {
			w.WriteHeader(httpErr.Status)
		}
		_ = bunrouter.JSON(w, httpErr)

		return err
	}
}

func corsMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
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
