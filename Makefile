migrate:
	go run migrations/auto.go
seed:
	go run cmd/seed/main.go
run:
	go run cmd/main.go
build:
	go build cmd/main.go