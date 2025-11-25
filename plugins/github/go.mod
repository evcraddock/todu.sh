module github.com/evcraddock/todu.sh/plugins/github

go 1.24.6

require (
	github.com/evcraddock/todu.sh v0.1.0
	github.com/google/go-github/v56 v56.0.0
	golang.org/x/oauth2 v0.33.0
)

require github.com/google/go-querystring v1.1.0 // indirect

replace github.com/evcraddock/todu.sh => ../..
