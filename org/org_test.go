package org_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uptrace/bun-realworld-app/org"
	"github.com/uptrace/bun-realworld-app/testbed"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func TestOrg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "org")
}

var _ = Describe("createUser", func() {
	var ctx context.Context
	var testapp *testbed.TestApp
	var data map[string]interface{}
	var userKeys Keys

	BeforeEach(func() {
		ctx = context.Background()
		testapp = testbed.StartApp(ctx)
		testapp.TruncateDB(ctx)

		userKeys = Keys{
			"username":  Equal("wangzitian0"),
			"email":     Equal("wzt@gg.cn"),
			"bio":       Equal("bar"),
			"image":     Equal("img"),
			"token":     Not(BeEmpty()),
			"following": Equal(false),
		}

		json := `{"user": {"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke", "image": "img", "bio": "bar"}}`
		resp := testapp.Client().PostJSON("/api/users", json)

		data = parseJSON(resp, http.StatusOK)
	})

	AfterEach(func() {
		testapp.Stop()
	})

	It("creates new user", func() {
		Expect(data["user"]).To(MatchAllKeys(userKeys))
	})

	Describe("loginUser", func() {
		var user *org.User

		BeforeEach(func() {
			json := `{"user": {"email": "wzt@gg.cn","password": "jakejxke"}}`
			resp := testapp.Client().PostJSON("/api/users/login", json)

			data = parseJSON(resp, http.StatusOK)

			username := data["user"].(map[string]interface{})["username"].(string)
			var err error
			user, err = org.SelectUserByUsername(ctx, testapp.App, username)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns user with JWT token", func() {
			Expect(data["user"]).To(MatchAllKeys(userKeys))
		})

		Describe("currentUser", func() {
			BeforeEach(func() {
				resp := testapp.Client().WithToken(user.ID).Get("/api/user/")
				data = parseJSON(resp, http.StatusOK)
			})

			It("returns logged in user", func() {
				Expect(data["user"]).To(MatchAllKeys(userKeys))
			})
		})

		Describe("updateUser", func() {
			BeforeEach(func() {
				json := `{"user": {"username": "hello","email": "foo@bar.com", "image": "bar", "bio": "foo"}}`
				resp := testapp.Client().WithToken(user.ID).PutJSON("/api/user/", json)
				data = parseJSON(resp, http.StatusOK)
			})

			It("returns updated user", func() {
				user := data["user"].(map[string]interface{})
				Expect(user).To(MatchAllKeys(Keys{
					"username":  Equal("hello"),
					"email":     Equal("foo@bar.com"),
					"bio":       Equal("foo"),
					"image":     Equal("bar"),
					"token":     Not(BeEmpty()),
					"following": Equal(false),
				}))
			})
		})

		Describe("followUser", func() {
			var username string

			BeforeEach(func() {
				json := `{"user": {"username": "hello","email": "foo@bar.com","password": "pwd"}}`
				resp := testapp.Client().PostJSON("/api/users", json)

				data = parseJSON(resp, http.StatusOK)

				username = data["user"].(map[string]interface{})["username"].(string)

				url := fmt.Sprintf("/api/profiles/%s/follow", username)
				resp = testapp.Client().WithToken(user.ID).Post(url, "")
				_ = parseJSON(resp, 200)

				url = fmt.Sprintf("/api/profiles/%s", username)
				resp = testapp.Client().WithToken(user.ID).Get(url)
				data = parseJSON(resp, 200)
			})

			It("returns followed profile", func() {
				profile := data["profile"].(map[string]interface{})
				Expect(profile).To(MatchAllKeys(Keys{
					"username":  Equal("hello"),
					"bio":       Equal(""),
					"image":     Equal(""),
					"following": Equal(true),
				}))
			})

			Describe("unfollowUser", func() {
				BeforeEach(func() {
					url := fmt.Sprintf("/api/profiles/%s/follow", username)
					resp := testapp.Client().WithToken(user.ID).Delete(url)
					data = parseJSON(resp, 200)
				})

				It("returns profile", func() {
					profile := data["profile"].(map[string]interface{})
					Expect(profile).To(MatchAllKeys(Keys{
						"username":  Equal("hello"),
						"bio":       Equal(""),
						"image":     Equal(""),
						"following": Equal(false),
					}))
				})
			})
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
