package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// client wraps the GitHub API client with configuration.
type client struct {
	gh  *github.Client
	ctx context.Context
}

// newClient creates a new GitHub API client.
//
// Configuration:
//   - token: GitHub personal access token (required)
//   - url: GitHub API URL (optional, defaults to public GitHub)
func newClient(config map[string]string) (*client, error) {
	token := config["token"]
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Create OAuth2 token source
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	var gh *github.Client
	if url := config["url"]; url != "" && url != "https://api.github.com" {
		// Use custom GitHub Enterprise URL
		var err error
		gh, err = github.NewClient(tc).WithEnterpriseURLs(url, url)
		if err != nil {
			return nil, fmt.Errorf("failed to create client with custom URL: %w", err)
		}
	} else {
		// Use public GitHub
		gh = github.NewClient(tc)
	}

	return &client{
		gh:  gh,
		ctx: ctx,
	}, nil
}

// listRepositories retrieves all repositories accessible to the authenticated user.
func (c *client) listRepositories(ctx context.Context) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.gh.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// getRepository retrieves a single repository by owner and name.
func (c *client) getRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	return repository, err
}

// listIssues retrieves issues for a repository.
func (c *client) listIssues(ctx context.Context, owner, repo string, since *time.Time) ([]*github.Issue, error) {
	var allIssues []*github.Issue
	opts := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		State:       "all",
	}

	if since != nil {
		opts.Since = *since
	}

	for {
		issues, resp, err := c.gh.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}
		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

// getIssue retrieves a single issue by number.
func (c *client) getIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issue, _, err := c.gh.Issues.Get(ctx, owner, repo, number)
	return issue, err
}

// createIssue creates a new issue in a repository.
func (c *client) createIssue(ctx context.Context, owner, repo string, req *github.IssueRequest) (*github.Issue, error) {
	issue, _, err := c.gh.Issues.Create(ctx, owner, repo, req)
	return issue, err
}

// updateIssue updates an existing issue.
func (c *client) updateIssue(ctx context.Context, owner, repo string, number int, req *github.IssueRequest) (*github.Issue, error) {
	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, req)
	return issue, err
}

// listComments retrieves all comments for an issue.
func (c *client) listComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
	var allComments []*github.IssueComment
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		comments, resp, err := c.gh.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, comments...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// createComment creates a new comment on an issue.
func (c *client) createComment(ctx context.Context, owner, repo string, number int, body string) (*github.IssueComment, error) {
	comment, _, err := c.gh.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: &body,
	})
	return comment, err
}
