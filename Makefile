run:
	go run cmd/app/main.go

build:
	go build -o load-balancer cmd/app/main.go

test:
	go test -v ./...

swag:
	@echo "Generating swagger docs.."
	swag init -d="./cmd/app,./internal/handler,./internal/model"
