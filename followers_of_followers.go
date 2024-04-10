package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v56/github"
)

func getFollowersOfFollowers(chFollowing chan *github.User, ctx context.Context, client *github.Client, username string, filter func(*github.User) bool) {
	for {
		followingList := getAllFollowingUsers(ctx, client, username)
		followersList := getAllFollowersUsers(ctx, client, username)

		followingSet := make(map[string]bool)
		for _, user := range followingList {
			followingSet[*user.Login] = true
		}

		for _, follower := range followersList {
			if filter(follower) {
				opts := &github.ListOptions{
					Page:    1,
					PerPage: 100,
				}
				followerFollowers, _, err := client.Users.ListFollowers(ctx, *follower.Login, opts)
				if err != nil {
					fmt.Printf("Error fetching follower list for %s: %v\n", *follower.Login, err)
					continue
				}

				for _, followerFollower := range followerFollowers {
					if filter(followerFollower) {

						if followingSet[*followerFollower.Login] {
							continue
						}

						followingSet[*followerFollower.Login] = true
						chFollowing <- followerFollower
					}
				}
			}
		}
	}
}
