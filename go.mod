module github.com/uptrace/bun-realworld-app

go 1.16

require (
	github.com/benbjohnson/clock v1.1.0
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-pg/urlstruct v1.0.1
	github.com/gosimple/slug v1.10.0
	github.com/jackc/pgx/v4 v4.13.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/uptrace/bun v1.0.9
	github.com/uptrace/bun/dbfixture v0.3.4
	github.com/uptrace/bun/dialect/pgdialect v1.0.9
	github.com/uptrace/bun/extra/bundebug v1.0.9
	github.com/uptrace/bunrouter v1.0.0-rc.1
	github.com/uptrace/bunrouter/extra/bunroutergzip v1.0.0-rc.1
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.0-rc.1
	github.com/uptrace/bunrouter/extra/reqlog v1.0.0-rc.1
	github.com/urfave/cli/v2 v2.3.0
	go4.org v0.0.0-20201209231011-d4a079459e60
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
