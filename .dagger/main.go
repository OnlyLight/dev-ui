package main

import (
	"context"
	"dagger/pipeline-dagger/internal/dagger"
	"fmt"
	"regexp"
)

type PipelineDagger struct {
	// +private
	Source *dagger.Directory
	// +private
	RepoFE string
	// +private
	Frontend *dagger.Frontend
}

// New initializes the pipeline with a Dagger client
func New(
	// +optional
	// +defaultPath="/"
	// +ignore=[".git", "**/node_modules"]
	source *dagger.Directory,
	// +optional
	// +default="github.com/OnlyLight/dev-ui"
	repoFE string,
) *PipelineDagger {
	return &PipelineDagger{
		Source:   source,
		RepoFE:   repoFE,
		Frontend: dag.Frontend(source.Directory("/website")),
	}
}

// BuildImage builds the Docker image using the existing Dockerfile
func (p *PipelineDagger) Build(
	ctx context.Context,
) *dagger.Container {
	return p.Frontend.Build()
}

// Run and debug the unit tests
func (p *PipelineDagger) Check(
	ctx context.Context,
	// Github token with permissions to comment on the pull request
	// +optional
	githubToken *dagger.Secret,
	// git commit in github
	// +optional
	commit string,
	// The model to use to debug debug tests
	// +optional
	model string,
) error {
	err := p.DebugUTIssues(ctx, githubToken, commit, model)
	return fmt.Errorf("Unit tests failed, attempting to debug %v", err)
}

// Publish the built image to a container registry
func (p *PipelineDagger) Publish(
	ctx context.Context,
	// +optional
	// +default="docker.io"
	dockerURL string,
	dockerUsername string,
	dockerPassword string,
	// +optional
	// +default="latest"
	tag string,
) (string, error) {
	container := p.Build(ctx)

	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`).MatchString(tag) {
		return "", fmt.Errorf("invalid tag format: %s", tag)
	}

	return container.
		WithRegistryAuth(dockerURL, dockerUsername, dag.SetSecret("docker-password", dockerPassword)).
		Publish(ctx, fmt.Sprintf("%s/crawler-website:%s", dockerUsername, tag))
}
