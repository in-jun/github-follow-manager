package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Print("GitHub 토큰을 입력하세요: ")
	var token string
	fmt.Scan(&token)

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	userInfo, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("Error fetching authenticated user information: %v\n", err)
		return
	}

	filter := func(user *github.User) bool {
		return true
	}

	followingChan := make(chan *github.User, 100)
	go getFollowersOfFollowers(followingChan, ctx, client, *userInfo.Login, filter)

	unfollowingChan := make(chan *github.User, 100)
	go getUnfollowingUsers(unfollowingChan, ctx, &wg, client, *userInfo.Login)

	go func() {
		for {
			unfollowedUser := <-unfollowingChan
			client.Users.Unfollow(ctx, *unfollowedUser.Login)
			fmt.Printf("Unfollowed user: %v\n", string(*unfollowedUser.Login))
			time.Sleep(time.Second)
		}
	}()

	for {
		select {
		case followedUser := <-followingChan:
			client.Users.Follow(ctx, *followedUser.Login)
			fmt.Printf("Followed user: %v\n", string(*followedUser.Login))
			time.Sleep(time.Minute)
		case <-sigs:
			wg.Wait()
			return
		}
	}
}
