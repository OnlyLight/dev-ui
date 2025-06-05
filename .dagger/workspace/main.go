package main

import (
	"context"
	"dagger/workspace/internal/dagger"
)

type Workspace struct {
	Work *dagger.Directory
	// +private
	Start *dagger.Directory
}

func New(
	// The source directory
	source *dagger.Directory,
) *Workspace {
	return &Workspace{Work: source, Start: source}
}

// Read a file in the Workspace
func (w *Workspace) ReadFile(
	ctx context.Context,
	// The path to the file in the workspace
	path string,
) (string, error) {
	return w.Work.File(path).Contents(ctx)
}

// Write a file to the Workspace
func (w *Workspace) WriteFile(
	// The path to the file in the workspace
	path string,
	// The new contents of the file
	contents string,
) *Workspace {
	w.Work = w.Work.WithNewFile(path, contents)
	return w
}

// Reset the workspace to the original state
func (w *Workspace) Reset() *Workspace {
	w.Work = w.Start
	return w
}

// List the files in the workspace in tree format
func (w *Workspace) Tree(ctx context.Context) (string, error) {
	return dag.Container().From("alpine:3").
		WithDirectory("/workspace", w.Work).
		WithExec([]string{"tree", "/workspace"}).
		Stdout(ctx)
}

// Show the changes made to the workspace so far in unified diff format
func (w *Workspace) Diff(ctx context.Context) (string, error) {
	return dag.Container().From("alpine:3").
		WithDirectory("/a", w.Start).
		WithDirectory("/b", w.Work).
		WithExec([]string{"diff", "-rN", "a/", "b/"}, dagger.ContainerWithExecOpts{Expect: dagger.ReturnTypeAny}).
		Stdout(ctx)
}

// Run the unit tests
func (w *Workspace) Test(ctx context.Context) (string, error) {
	return dag.Container().Build(w.Work, dagger.ContainerBuildOpts{
		Target: "build",
	}).WithEnvVariable("CI", "true").WithExec([]string{"npm", "test"}).Stdout(ctx)
}
