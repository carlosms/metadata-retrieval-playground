package subcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-cli.v0"
)

type LimitsCommand struct {
	cli.Command `name:"limits" short-description:"" long-description:""`
	Token       string `long:"token" short:"t" env:"SOURCED_GITHUB_TOKEN" description:"GitHub personal access token" required:"true"`
}

func (c *LimitsCommand) Execute(args []string) error {
	httpClient := oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	))

	v3Client := github.NewClient(httpClient)

	limit, _, err := v3Client.RateLimits(context.TODO())
	if err != nil {
		return err
	}

	fmt.Println("v3 limits:")
	fmt.Println(prettyPrint(limit))

	v4Client := githubv4.NewClient(httpClient)

	var q struct {
		RateLimit struct {
			Limit     int
			Remaining int
			ResetAt   time.Time
		}
	}

	err = v4Client.Query(context.TODO(), &q, nil)
	if err != nil {
		return err
	}

	fmt.Println("\nv4 limits:")
	fmt.Println(prettyPrint(q))
	return nil
}

func prettyPrint(v interface{}) string {
	s, _ := json.MarshalIndent(v, "", "  ")
	return string(s)
}
