build:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/gitlab-mcp ./main.go

build-cli:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/gitlab-cli ./cmd/gitlab-cli/

dev:
  go run main.go --env .env --sse_port 3001

install:
  go install ./...

install-cli:
  CGO_ENABLED=0 go build -ldflags="-s -w" -o $(go env GOPATH)/bin/gitlab-cli ./cmd/gitlab-cli/
