package github

import (
	"context"
	"fmt"

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
