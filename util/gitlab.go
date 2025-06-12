package util

import (
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var GitlabClient = sync.OnceValue[*gitlab.Client](func() *gitlab.Client {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		log.Fatal("GITLAB_TOKEN is required")
	}

	host := os.Getenv("GITLAB_URL")
	if host == "" {
		log.Fatal("GITLAB_URL is required")
	}

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(host))
	if err != nil {
		log.Fatal(errors.WithMessage(err, "failed to create gitlab client"))
	}

	return client
})