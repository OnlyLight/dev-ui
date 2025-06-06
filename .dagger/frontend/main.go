package main

import (
	"context"
	"dagger/frontend/internal/dagger"
)

type Frontend struct {
	Source *dagger.Directory
}

func New(source *dagger.Directory) *Frontend {
	return &Frontend{
		Source: source,
	}
}

// Run the unit tests
func (f *Frontend) UnitTest(ctx context.Context) (string, error) {
	return dag.Container().Build(f.Source, dagger.ContainerBuildOpts{
		Target: "build",
	}).WithEnvVariable("CI", "true").WithExec([]string{"npm", "test"}).Stdout(ctx)
}

// BuildImage builds the Docker image using the existing Dockerfile
func (f *Frontend) Build(
	ctx context.Context,
) *dagger.Container {
	return dag.Container().Build(f.Source)
}

// Stateless checker
func (f *Frontend) CheckDirectory(
	ctx context.Context,
	// Directory to run checks on
	source *dagger.Directory,
) (string, error) {
	f.Source = source
	return f.UnitTest(ctx)
}
