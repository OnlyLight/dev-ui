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
		WithWorkspaceOutput("fixed", "workspace with fixed tests")

	// Put it all together to form the agent (LLM agent that fixes the tests)
	work := dag.LLM(dagger.LLMOpts{Model: model}).
		WithEnv(environment).
		WithPromptFile(prompt)

	// Bind the LLM's output to the workspace output
	environment = work.Env()

	// Get output from the agent and return the diff
	return environment.Output("fixed").AsWorkspace().Diff(ctx)
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

	pr, err := gh.GetPrForCommit(ctx, p.RepoFE, commit)
	if err != nil {
		return fmt.Errorf("failed to get PR for commit: %w", err)
	}

	// Suggest fix
	suggestionDiff, err := p.DebugUT(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to debug UT: %w", err)
	}
	if suggestionDiff == "" {
		return fmt.Errorf("no suggestions found")
	}

	fmt.Printf("Raw diff content:\n%s\n", suggestionDiff)

	// Convert the diff to CodeSuggestions
	codeSuggestions, err := parseDiff(suggestionDiff)
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	fmt.Printf("Number of suggestions: %d\n", len(codeSuggestions))

	// For each suggestion, comment on PR
	for i, suggestion := range codeSuggestions {
		// Ensure we have a valid file path (remove any 'b/' prefix that git diff adds)
		filePath := strings.TrimPrefix(suggestion.File, "b/")

		fmt.Printf("\nSuggestion %d:\n", i+1)
		fmt.Printf("File: %s\n", filePath)
		fmt.Printf("Line: %d\n", suggestion.Line)
		fmt.Printf("DiffHunk:\n%s\n", suggestion.DiffHunk)
		fmt.Printf("Suggestion:\n%s\n", strings.Join(suggestion.Suggestion, "\n"))

		// Create the comment with the required diff_hunk
		comment := fmt.Sprintf("```diff\n%s\n```\n\n```suggestion\n%s\n```",
			suggestion.DiffHunk,
			strings.Join(suggestion.Suggestion, "\n"))

		err := gh.WritePullRequestCodeComment(
			ctx,
			p.RepoFE,
			pr,
			commit,
			comment,
			filePath,
			"RIGHT",
			suggestion.Line)
		if err != nil {
			return fmt.Errorf("failed to write PR comment for file %s at line %d: %w", filePath, suggestion.Line, err)
		}
	}
	return nil
}
