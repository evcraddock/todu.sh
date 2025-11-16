package main

import (
	"github.com/evcraddock/todu.sh/cmd/todu/cmd"
	_ "github.com/evcraddock/todu.sh/plugins/forgejo" // Register Forgejo plugin
	_ "github.com/evcraddock/todu.sh/plugins/github"  // Register GitHub plugin
	_ "github.com/evcraddock/todu.sh/plugins/todoist" // Register Todoist plugin
)

func main() {
	cmd.Execute()
}
