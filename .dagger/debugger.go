package main

import (
	"context"
	"dagger/pipeline-dagger/internal/dagger"
	"fmt"
	"strings"
)

// Debug broken tests. Returns a unified diff of the test fixes
func (p *PipelineDagger) DebugUT(
	ctx context.Context,
	// The model to use to debug debug tests
	// +optional
	// +default = "gemini-2.0-flash"
	model string,
) (string, error) {
	// Detailed prompt stored in markdown file
	prompt := dag.CurrentModule().Source().File("prompts/fix_test.md")

	// Environment with agent inputs and outputs
	environment := dag.Env().
		WithWorkspaceInput("workspace", dag.Workspace(p.Source), "workspace to read, write, and test code").
		WithWorkspaceOutput("successful", "the workspace with successful tests")

	// Put it all together to form the agent (LLM agent that fixes the tests)
	work := dag.LLM(dagger.LLMOpts{Model: model}).
		WithEnv(environment).
		WithPromptFile(prompt)

	// Get output from the agent and return the diff
	return work.Env().Output("successful").AsWorkspace().Diff(ctx)
}

// Suggest fixes to a Issues
func (p *PipelineDagger) DebugUTIssues(
	ctx context.Context,
	// Github Token with permissions to write issues and contents
	githubToken *dagger.Secret,
	// Git commit in Github
	commit string,
	// The model to use to debug debug tests
	// +optional
	// +default = "gemini-2.0-flash"
	model string,
) error {
	gh := dag.GithubIssue(dagger.GithubIssueOpts{Token: githubToken})

	gitRef := dag.Git(p.RepoFE).Commit(commit)
	gitSource := gitRef.Tree()
	pr, err := gh.GetPrForCommit(ctx, p.RepoFE, commit)
	if err != nil {
		return err
	}

	// Set source to PR head
	p = New(gitSource, p.RepoFE, p.RepoInfra)

	// Suggest fix
	suggestionDiff, err := p.DebugUT(ctx, model)
	if err != nil {
		return err
	}
	if suggestionDiff == "" {
		return fmt.Errorf("no suggestions found")
	}

	// Convert the diff to CodeSuggestions
	codeSuggestions := parseDiff(suggestionDiff)

	// For each suggestion, comment on PR
	for _, suggestion := range codeSuggestions {
		markupSuggestion := "```suggestion\n" + strings.Join(suggestion.Suggestion, "\n") + "\n```"
		err := gh.WritePullRequestCodeComment(
			ctx,
			p.RepoFE,
			pr,
			commit,
			markupSuggestion,
			suggestion.File,
			"RIGHT",
			suggestion.Line)
		if err != nil {
			return err
		}
	}
	return nil
}
