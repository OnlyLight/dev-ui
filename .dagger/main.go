package main

import (
	"context"
	"dagger/pipeline-dagger/internal/dagger"
	"fmt"
	"regexp"
)

type PipelineDagger struct {
	Source *dagger.Directory // Shared Dagger directory
}

// New initializes the pipeline with a Dagger client
func New(
	// +optional
	// +defaultPath="/website"
	// +ignore=[".git", "**/node_modules"]
	source *dagger.Directory,
) *PipelineDagger {
	return &PipelineDagger{Source: source}
}

// BuildImage builds the Docker image using the existing Dockerfile
func (p *PipelineDagger) Build(
	ctx context.Context,
) *dagger.Container {
	return dag.Container().Build(p.Source, dagger.ContainerBuildOpts{
		Dockerfile: "Dockerfile",
	})
}

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
