package testbed

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/uptrace/bun-realworld-app/app"
)

func ExtendKeys(a, b gstruct.Keys) gstruct.Keys {
	res := make(gstruct.Keys)
	for k, v := range a {
		res[k] = v
	}
	for k, v := range b {
		res[k] = v
	}
	return res
}

func ResetAll(ctx context.Context) {
	truncateDB(ctx)
}

func truncateDB(ctx context.Context) {
	cmd := "TRUNCATE users, favorite_articles, follow_users, comments, articles, article_tags"
	_, err := app.DB().ExecContext(ctx, cmd)
	Expect(err).NotTo(HaveOccurred())
}
