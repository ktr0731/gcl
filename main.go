package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v28/github"
	"github.com/morikuni/failure"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var eg errgroup.Group
	for _, a := range os.Args[1:] {
		a := a
		eg.Go(func() error {
			sp := strings.Split(a, "/")
			if len(sp) != 2 {
				return failure.Unexpected("invalid format", failure.Context{"input": a})
			}

			owner, repo := sp[0], sp[1]

			repository, _, err := client.Repositories.Get(ctx, owner, repo)
			if err != nil {
				return failure.Wrap(err)
			}
			release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
			if err != nil {
				return failure.Wrap(err)
			}
			if release.CreatedAt.Before(repository.GetUpdatedAt().Time) {
				fmt.Println(strings.Join([]string{"github.com", owner, repo}, "/"))
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "ghl: %s", err)
		os.Exit(1)
	}
}
