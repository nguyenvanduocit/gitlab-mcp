build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/gitlab-mcp ./main.go

dev:
  go run main.go --env .env --sse_port 3001

install:
  go install ./...
