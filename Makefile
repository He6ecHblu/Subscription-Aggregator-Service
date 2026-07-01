.PHONY: test test-integration docker-up docker-down

DATABASE_URL ?= postgres://subscriptions_user:subscriptions_password@localhost:5432/subscriptions_db?sslmode=disable

test:
	go test ./...

test-integration:
	DATABASE_URL="$(DATABASE_URL)" go test ./internal/repository -run TestSubscriptionRepositoryIntegration -count=1 -v

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
