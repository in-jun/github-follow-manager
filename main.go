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

	followingChan := make(chan *github.User, 10)
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
			time.Sleep(time.Minute * 5)
		case <-sigs:
			wg.Wait()
			return
		}
	}
}

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
					if !followingSet[*followerFollower.Login] && filter(followerFollower) {
						followingSet[*followerFollower.Login] = true
						chFollowing <- followerFollower
					}
				}
			}
		}
	}
}

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

var checkedUsers = make(map[string]bool)
var checkedUsersMutex sync.Mutex

func getUnfollowingUsers(chUnfollowing chan *github.User, ctx context.Context, wg *sync.WaitGroup, client *github.Client, username string) {
	for {
		followerMap := make(map[string]bool)
		followingList := getAllFollowingUsers(ctx, client, username)
		followersList := getAllFollowersUsers(ctx, client, username)

		for _, follower := range followersList {
			followerMap[*follower.Login] = true
		}

		for _, followingUser := range followingList {
			if !followerMap[*followingUser.Login] && !isCheckedUser(*followingUser.Login) {
				setCheckedUser(*followingUser.Login, true)
				wg.Add(1)
				go func(user *github.User) {
					defer wg.Done()
					sigs := make(chan os.Signal, 1)
					signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
					select {
					case <-time.After(24 * time.Hour):
						if !followerMap[*user.Login] {
							chUnfollowing <- user
						}
						setCheckedUser(*user.Login, false)
					case <-sigs:
						if !followerMap[*user.Login] {
							chUnfollowing <- user
						}
						setCheckedUser(*user.Login, false)
					}
				}(followingUser)
			}
		}
		time.Sleep(time.Second * 30)
	}
}

func setCheckedUser(login string, value bool) {
	checkedUsersMutex.Lock()
	defer checkedUsersMutex.Unlock()
	checkedUsers[login] = value
}

func isCheckedUser(login string) bool {
	checkedUsersMutex.Lock()
	defer checkedUsersMutex.Unlock()
	return checkedUsers[login]
}
