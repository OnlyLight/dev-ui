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

	ws := dag.Workspace(
		p.Frontend.Source(),
		p.Frontend.AsWorkspaceCheckable(),
	)

	// Environment with agent inputs and outputs
	environment := dag.Env().
		WithWorkspaceInput("workspace", ws, "workspace to read, write, and test code").
		WithWorkspaceOutput("fixed", "workspace with fixed tests")

	// Put it all together to form the agent (LLM agent that fixes the tests)
	return dag.LLM(dagger.LLMOpts{Model: model}).
		WithEnv(environment).
		WithPromptFile(prompt).
		Env(). // Bind the LLM's output to the workspace output
		Output("fixed").
		AsWorkspace().
		Diff(ctx) // Get output from the agent and return the diff
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

	// Determine PR head
	gitRef := dag.Git(p.RepoFE).Commit(commit)
	gitSource := gitRef.Tree()
	pr, err := gh.GetPrForCommit(ctx, p.RepoFE, commit)
	if err != nil {
		return fmt.Errorf("failed to get PR for commit: %w", err)
	}

	// Set source to PR head
	p = New(gitSource, p.RepoFE)

	// Suggest fix
	suggestionDiff, err := p.DebugUT(ctx, model)
	if err != nil {
		return fmt.Errorf("debug UT failed: %w", err)
	}
	if strings.TrimSpace(suggestionDiff) == "" {
		return nil
	}

	fmt.Printf("Raw diff content:\n%s\n", suggestionDiff)

	// Convert the diff to CodeSuggestions
	codeSuggestions := parseDiff(suggestionDiff)

	fmt.Printf("Number of suggestions: %d\n", len(codeSuggestions))

	// For each suggestion, comment on PR
	for i, suggestion := range codeSuggestions {
		fmt.Printf("Suggestion %d: File=%s, Line=%d\n", i+1, suggestion.File, suggestion.Line)
		if suggestion.File == "" {
			return fmt.Errorf("invalid suggestion: empty file path")
		}
		if suggestion.Line < 1 {
			return fmt.Errorf("invalid suggestion: line %d in %s", suggestion.Line, suggestion.File)
		}

		markupSuggestion := "```suggestion\n" + strings.Join(suggestion.Suggestion, "\n") + "\n```"
		err := gh.WritePullRequestCodeComment(
			ctx,
			p.RepoFE,
			pr,
			commit,
			markupSuggestion,
			suggestion.File,
			"RIGHT",
			suggestion.Line,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
