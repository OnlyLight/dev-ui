package main

import (
	"context"
	"dagger/workspace/internal/dagger"
)

// Interface for something that can be checked
type Checkable interface {
	dagger.DaggerObject
	CheckDirectory(ctx context.Context, source *dagger.Directory) (string, error)
}

type Workspace struct {
	Work *dagger.Directory
	// +private
	Start *dagger.Directory
	// +private
	Checker Checkable
}

func New(
	// The source directory
	source *dagger.Directory,
	// Checker to use for testing
	checker Checkable,
) *Workspace {
	return &Workspace{Work: source, Start: source, Checker: checker}
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

// Run the tests in the workspace
func (w *Workspace) Check(ctx context.Context) (string, error) {
	return w.Checker.CheckDirectory(ctx, w.Work)
}
