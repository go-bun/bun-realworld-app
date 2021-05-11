db_reset:
	sudo -u postgres psql -c "DROP DATABASE IF EXISTS realworld_test"
	sudo -u postgres psql -c "CREATE DATABASE realworld_test"

	make db_migrate

db_migrate:
	go run cmd/bun/main.go -env=test db init
	go run cmd/bun/main.go -env=test db migrate

test:
	TZ= go test ./org
	TZ= go test ./blog

api_test:
	TZ= go run cmd/bun/main.go -env=test api &
	APIURL=http://localhost:8000/api ./scripts/run-api-tests.sh
