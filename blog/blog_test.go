package blog_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/uptrace/bun-realworld-app/org"
	"github.com/uptrace/bun-realworld-app/testbed"
	"github.com/uptrace/bun/dbfixture"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "blog")
}

var _ = Describe("createArticle", func() {
	var ctx context.Context
	var app *testbed.TestApp

	var userClient testbed.Client

	var data map[string]interface{}
	var slug string
	var user *org.User

	var helloArticleKeys, fooArticleKeys, favoritedArticleKeys Keys

	createFollowedUser := func() *org.User {
		followedUser := &org.User{
			Username:     "FollowedUser",
			Email:        "foo@bar.com",
			PasswordHash: "h2",
		}
		_, err := app.DB().NewInsert().Model(followedUser).Exec(ctx)
		Expect(err).NotTo(HaveOccurred())

		url := fmt.Sprintf("/api/profiles/%s/follow", followedUser.Username)
		resp := userClient.Post(url, "")
		_ = parseJSON(resp, 200)

		return followedUser
	}

	BeforeEach(func() {
		ctx = context.Background()
		app = testbed.StartApp(ctx)
		app.TruncateDB(ctx)

		helloArticleKeys = Keys{
			"title":          Equal("Hello world"),
			"slug":           HavePrefix("hello-world-"),
			"description":    Equal("Hello world article description!"),
			"body":           Equal("Hello world article body."),
			"author":         Equal(map[string]interface{}{"following": false, "username": "CurrentUser", "bio": "", "image": ""}),
			"tagList":        ConsistOf([]interface{}{"greeting", "welcome", "salut"}),
			"favoritesCount": Equal(float64(0)),
			"favorited":      Equal(false),
			"createdAt":      Equal(app.Clock().Now().Format(time.RFC3339Nano)),
			"updatedAt":      Equal(app.Clock().Now().Format(time.RFC3339Nano)),
		}

		favoritedArticleKeys = testbed.ExtendKeys(helloArticleKeys, Keys{
			"favorited":      Equal(true),
			"favoritesCount": Equal(float64(1)),
		})

		fooArticleKeys = Keys{
			"title":          Equal("Foo bar"),
			"slug":           HavePrefix("foo-bar-"),
			"description":    Equal("Foo bar article description!"),
			"body":           Equal("Foo bar article body."),
			"author":         Equal(map[string]interface{}{"following": false, "username": "CurrentUser", "bio": "", "image": ""}),
			"tagList":        ConsistOf([]interface{}{"foobar", "variable"}),
			"favoritesCount": Equal(float64(0)),
			"favorited":      Equal(false),
			"createdAt":      Equal(app.Clock().Now().Format(time.RFC3339Nano)),
			"updatedAt":      Equal(app.Clock().Now().Format(time.RFC3339Nano)),
		}

		app.DB().RegisterModel((*org.User)(nil))

		fixture := dbfixture.New(app.DB())
		err := fixture.Load(ctx, os.DirFS("testdata"), "fixture.yaml")
		Expect(err).NotTo(HaveOccurred())

		user = fixture.MustRow("User.current").(*org.User)
		userClient = app.Client().WithToken(user.ID)
	})

	BeforeEach(func() {
		json := `{"article": {"title": "Hello world", "description": "Hello world article description!", "body": "Hello world article body.", "tagList": ["greeting", "welcome", "salut"]}}`
		resp := userClient.PostJSON("/api/articles", json)

		data = parseJSON(resp, http.StatusOK)
		slug = data["article"].(map[string]interface{})["slug"].(string)
	})

	It("creates new article", func() {
		Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
	})

	Describe("showFeed", func() {
		BeforeEach(func() {
			followedUser := createFollowedUser()

			json := `{"article": {"title": "Foo bar", "description": "Foo bar article description!", "body": "Foo bar article body.", "tagList": ["foobar", "variable"]}}`
			resp := app.Client().WithToken(followedUser.ID).PostJSON("/api/articles", json)

			_ = parseJSON(resp, http.StatusOK)

			resp = userClient.Get("/api/articles/feed")
			data = parseJSON(resp, http.StatusOK)
		})

		It("returns article", func() {
			articles := data["articles"].([]interface{})

			Expect(articles).To(HaveLen(1))
			followedAuthorKeys := testbed.ExtendKeys(fooArticleKeys, Keys{
				"author": Equal(map[string]interface{}{
					"following": true,
					"username":  "FollowedUser",
					"bio":       "",
					"image":     "",
				}),
			})
			Expect(articles[0].(map[string]interface{})).To(MatchAllKeys(followedAuthorKeys))
		})
	})

	Describe("showArticle", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s", slug)
			resp := app.Client().Get(url)

			data = parseJSON(resp, http.StatusOK)
		})

		It("returns article", func() {
			Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
		})
	})

	Describe("listArticles", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s?author=CurrentUser", slug)
			resp := app.Client().Get(url)

			data = parseJSON(resp, http.StatusOK)
		})

		It("returns articles by author", func() {
			Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
		})
	})

	Describe("favoriteArticle", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s/favorite", slug)
			resp := userClient.Post(url, "")
			_ = parseJSON(resp, 200)

			url = fmt.Sprintf("/api/articles/%s", slug)
			resp = userClient.Get(url)
			data = parseJSON(resp, 200)
		})

		It("returns favorited article", func() {
			Expect(data["article"]).To(MatchAllKeys(favoritedArticleKeys))
		})

		Describe("unfavoriteArticle", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/favorite", slug)
				resp := userClient.Delete(url)
				_ = parseJSON(resp, 200)

				url = fmt.Sprintf("/api/articles/%s", slug)
				resp = userClient.Get(url)
				data = parseJSON(resp, 200)
			})

			It("returns article", func() {
				Expect(data["article"]).To(MatchAllKeys(helloArticleKeys))
			})
		})
	})

	Describe("listArticles", func() {
		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s/favorite", slug)
			resp := userClient.Post(url, "")
			_ = parseJSON(resp, 200)

			resp = userClient.Get("/api/articles")
			data = parseJSON(resp, 200)
		})

		It("returns articles", func() {
			articles := data["articles"].([]interface{})

			Expect(articles).To(HaveLen(1))
			article := articles[0].(map[string]interface{})
			Expect(article).To(MatchAllKeys(favoritedArticleKeys))
		})
	})

	Describe("updateArticle", func() {
		BeforeEach(func() {
			json := `{"article": {"title": "Foo bar", "description": "Foo bar article description!", "body": "Foo bar article body.", "tagList": []}}`

			url := fmt.Sprintf("/api/articles/%s", slug)
			resp := userClient.PutJSON(url, json)
			data = parseJSON(resp, 200)
		})

		It("returns article", func() {
			updatedArticleKeys := testbed.ExtendKeys(fooArticleKeys, Keys{
				"slug":      HavePrefix("hello-world-"),
				"tagList":   Equal([]interface{}{}),
				"updatedAt": Equal(app.Clock().Now().Format(time.RFC3339Nano)),
			})
			Expect(data["article"]).To(MatchAllKeys(updatedArticleKeys))
		})
	})

	Describe("deleteArticle", func() {
		var resp *httptest.ResponseRecorder

		BeforeEach(func() {
			url := fmt.Sprintf("/api/articles/%s", slug)
			resp = userClient.Delete(url)
		})

		It("deletes article", func() {
			Expect(resp.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("createComment", func() {
		var commentKeys Keys
		var commentID uint64
		var followedUser *org.User

		BeforeEach(func() {
			commentKeys = Keys{
				"id":        Not(BeZero()),
				"body":      Equal("First comment."),
				"author":    Equal(map[string]interface{}{"following": false, "username": "FollowedUser", "bio": "", "image": ""}),
				"createdAt": Equal(app.Clock().Now().Format(time.RFC3339Nano)),
				"updatedAt": Equal(app.Clock().Now().Format(time.RFC3339Nano)),
			}

			followedUser = createFollowedUser()

			json := `{"comment": {"body": "First comment."}}`
			url := fmt.Sprintf("/api/articles/%s/comments", slug)
			resp := app.Client().WithToken(followedUser.ID).PostJSON(url, json)
			data = parseJSON(resp, 200)

			commentID = uint64(data["comment"].(map[string]interface{})["id"].(float64))
		})

		It("returns created comment to article", func() {
			Expect(data["comment"]).To(MatchAllKeys(commentKeys))
		})

		Describe("showComment", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp := app.Client().Get(url)
				data = parseJSON(resp, 200)
			})

			It("returns article comments", func() {
				Expect(data["comment"]).To(MatchAllKeys(commentKeys))
			})
		})

		Describe("showComment with authentication", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp := userClient.Get(url)
				data = parseJSON(resp, 200)
			})

			It("returns article comments", func() {
				followedCommentKeys := testbed.ExtendKeys(commentKeys, Keys{
					"author": Equal(map[string]interface{}{"following": true, "username": "FollowedUser", "bio": "", "image": ""}),
				})
				Expect(data["comment"]).To(MatchAllKeys(followedCommentKeys))
			})
		})

		Describe("listArticleComments", func() {
			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments", slug)
				resp := userClient.Get(url)
				data = parseJSON(resp, 200)
			})

			It("returns article comments", func() {
				followedCommentKeys := testbed.ExtendKeys(commentKeys, Keys{
					"author": Equal(map[string]interface{}{"following": true, "username": "FollowedUser", "bio": "", "image": ""}),
				})
				Expect(data["comments"].([]interface{})[0]).To(MatchAllKeys(followedCommentKeys))
			})
		})

		Describe("deleteComment", func() {
			var resp *httptest.ResponseRecorder

			BeforeEach(func() {
				url := fmt.Sprintf("/api/articles/%s/comments/%d", slug, commentID)
				resp = app.Client().WithToken(followedUser.ID).Delete(url)
			})

			It("deletes comment", func() {
				Expect(resp.Code).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("listTags", func() {
		BeforeEach(func() {
			resp := app.Client().Get("/api/tags/")
			data = parseJSON(resp, 200)
		})

		It("returns tags", func() {
			Expect(data["tags"]).To(ConsistOf([]string{
				"greeting",
				"salut",
				"welcome",
			}))
		})
	})
})

func parseJSON(resp *httptest.ResponseRecorder, code int) map[string]interface{} {
	out := make(map[string]interface{})
	err := json.Unmarshal(resp.Body.Bytes(), &out)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.Code).To(Equal(code))
	return out
}
