package bunapp

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
)

func setupUptrace(app *App) {
	uptrace.ConfigureOpentelemetry(&uptrace.Config{
		DSN:            app.Config().Uptrace.DSN,
		ServiceName:    "api",
		ServiceVersion: "1.0.0",
	})

	app.OnAfterStop("uptrace.Shutdown", func(ctx context.Context, _ *App) error {
		return uptrace.Shutdown(ctx)
	})
}
