package main

import (
	"context"

	"github.com/google/go-github/v56/github"
)

func getAllFollowersUsers(ctx context.Context, client *github.Client, username string) []*github.User {
	var allFollowers []*github.User

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		followers, resp, err := client.Users.ListFollowers(ctx, username, opts)
		if err != nil {
			return nil
		}

		allFollowers = append(allFollowers, followers...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFollowers
}

func getAllFollowingUsers(ctx context.Context, client *github.Client, username string) []*github.User {
	var allFollowing []*github.User

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		following, resp, err := client.Users.ListFollowing(ctx, username, opts)
		if err != nil {
			return nil
		}

		allFollowing = append(allFollowing, following...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFollowing
}
