package app

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/vmihailenco/treemux"
)

type App struct {
	ctx context.Context
	cfg *AppConfig

	stopping uint32
	stopCh   chan struct{}

	clock clock.Clock

	router    *treemux.Router
	apiRouter *treemux.Group

	// lazy init
	dbOnce sync.Once
	db     *bun.DB
}

func New(ctx context.Context, cfg *AppConfig) *App {
	app := &App{
		ctx:    ctx,
		cfg:    cfg,
		stopCh: make(chan struct{}),
		clock:  clock.New(),
	}
	app.initRouter()
	return app
}

func Start(ctx context.Context, service, envName string) (*App, error) {
	cfg, err := ReadConfig(service, envName)
	if err != nil {
		return nil, err
	}
	return StartConfig(ctx, cfg)
}

func StartConfig(ctx context.Context, cfg *AppConfig) (*App, error) {
	rand.Seed(time.Now().UnixNano())

	app := New(ctx, cfg)
	if err := onStart.Run(ctx, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *App) Stop() {
	_ = onStop.Run(app.ctx, app)
	_ = onAfterStop.Run(app.ctx, app)
}

func (app *App) Context() context.Context {
	return app.ctx
}

func (app *App) Config() *AppConfig {
	return app.cfg
}

func (app *App) Running() bool {
	return !app.Stopping()
}

func (app *App) Stopping() bool {
	return atomic.LoadUint32(&app.stopping) == 1
}

func (app *App) IsDebug() bool {
	return app.cfg.Debug
}

func (app *App) Clock() clock.Clock {
	return app.clock
}

func (app *App) SetClock(clock clock.Clock) {
	app.clock = clock
}

func (app *App) Router() *treemux.Router {
	return app.router
}

func (app *App) APIRouter() *treemux.Group {
	return app.apiRouter
}

func (app *App) DB() *bun.DB {
	app.dbOnce.Do(func() {
		config, err := pgx.ParseConfig(app.cfg.PGX.DSN)
		if err != nil {
			panic(err)
		}

		config.PreferSimpleProtocol = true
		sqldb := stdlib.OpenDB(*config)

		db := bun.Open(sqldb, pgdialect.New())
		// db.AddQueryHook(pgotel.TracingHook{})
		if app.IsDebug() {
			db.AddQueryHook(bundebug.NewQueryHook())
		}

		app.db = db
	})
	return app.db
}

func WaitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}
