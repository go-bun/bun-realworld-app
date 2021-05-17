package testbed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/uptrace/bun-realworld-app/bunapp"
	"github.com/uptrace/bun-realworld-app/org"
)

type TestApp struct {
	*bunapp.App
}

func StartApp(ctx context.Context) *TestApp {
	_, app, err := bunapp.Start(ctx, "test", "test")
	if err != nil {
		panic(err)
	}

	mock := clock.NewMock()
	mock.Set(time.Date(2020, time.January, 1, 2, 3, 4, 5000, time.UTC))
	app.SetClock(mock)

	return &TestApp{
		App: app,
	}
}

func (app *TestApp) Client() Client {
	return Client{
		app: app.App,
	}
}

func (app *TestApp) TruncateDB(ctx context.Context) {
	query := "TRUNCATE users, favorite_articles, follow_users, comments, articles, article_tags"
	_, err := app.DB().ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
}

//------------------------------------------------------------------------------

type Client struct {
	app *bunapp.App

	userID uint64
}

func (c Client) WithToken(userID uint64) Client {
	return Client{
		app:    c.app,
		userID: userID,
	}
}

func (c Client) Get(url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", url, nil)
	return c.Serve(req)
}

func (c Client) Post(url string, s string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, strings.NewReader(s))
	return c.Serve(req)
}

func (c Client) PostJSON(url string, s string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", url, strings.NewReader(s))
	req.Header.Set("Content-Type", "application/json")
	return c.Serve(req)
}

func (c Client) PutJSON(url string, s string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PUT", url, strings.NewReader(s))
	req.Header.Set("Content-Type", "application/json")
	return c.Serve(req)
}

func (c Client) Delete(url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", url, nil)
	return c.Serve(req)
}

func (c Client) Serve(req *http.Request) *httptest.ResponseRecorder {
	if c.userID != 0 {
		token, err := org.CreateUserToken(c.app, c.userID, time.Hour)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Authorization", "Token "+token)
	}

	resp := httptest.NewRecorder()
	c.app.Router().ServeHTTP(resp, req)
	return resp
}
