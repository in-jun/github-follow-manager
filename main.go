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

type GitHubManager struct {
	client       *github.Client
	ctx          context.Context
	username     string
	wg           *sync.WaitGroup
	checkedUsers map[string]bool
	usersMutex   sync.RWMutex
}

func NewGitHubManager(token string) (*GitHubManager, error) {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))

	userInfo, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error fetching user info: %v", err)
	}

	return &GitHubManager{
		client:       client,
		ctx:          ctx,
		username:     *userInfo.Login,
		wg:           &sync.WaitGroup{},
		checkedUsers: make(map[string]bool),
	}, nil
}

func (gm *GitHubManager) Run() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	followingChan := make(chan *github.User, 10)
	unfollowingChan := make(chan *github.User, 100)

	go gm.handleFollowing(followingChan)
	go gm.handleUnfollowing(unfollowingChan)
	go gm.processUnfollowing(unfollowingChan)

	<-sigs
	gm.wg.Wait()
}

func (gm *GitHubManager) handleFollowing(followingChan chan *github.User) {
	for {
		followingList := gm.getAllUsers(gm.client.Users.ListFollowing)
		followersList := gm.getAllUsers(gm.client.Users.ListFollowers)

		followingSet := make(map[string]bool)
		for _, user := range followingList {
			followingSet[*user.Login] = true
		}

		for _, follower := range followersList {
			opts := &github.ListOptions{PerPage: 100}
			followers, _, err := gm.client.Users.ListFollowers(gm.ctx, *follower.Login, opts)
			if err != nil {
				continue
			}

			for _, f := range followers {
				if !followingSet[*f.Login] {
					followingSet[*f.Login] = true
					followingChan <- f
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

func (gm *GitHubManager) handleUnfollowing(unfollowingChan chan *github.User) {
	for {
		followerMap := make(map[string]bool)
		followingList := gm.getAllUsers(gm.client.Users.ListFollowing)
		followersList := gm.getAllUsers(gm.client.Users.ListFollowers)

		for _, follower := range followersList {
			followerMap[*follower.Login] = true
		}

		for _, following := range followingList {
			if !followerMap[*following.Login] && !gm.isCheckedUser(*following.Login) {
				gm.setCheckedUser(*following.Login, true)
				gm.wg.Add(1)
				go gm.checkUserAfterDelay(following, followerMap, unfollowingChan)
			}
		}
		time.Sleep(30 * time.Second)
	}
}

func (gm *GitHubManager) processUnfollowing(unfollowingChan chan *github.User) {
	for user := range unfollowingChan {
		gm.client.Users.Unfollow(gm.ctx, *user.Login)
		fmt.Printf("Unfollowed user: %v\n", *user.Login)
		time.Sleep(time.Second)
	}
}

func (gm *GitHubManager) checkUserAfterDelay(user *github.User, followerMap map[string]bool, unfollowingChan chan *github.User) {
	defer gm.wg.Done()

	select {
	case <-time.After(24 * time.Hour):
		if !followerMap[*user.Login] {
			unfollowingChan <- user
		}
	case <-gm.ctx.Done():
		return
	}
	gm.setCheckedUser(*user.Login, false)
}

func (gm *GitHubManager) getAllUsers(listFunc func(context.Context, string, *github.ListOptions) ([]*github.User, *github.Response, error)) []*github.User {
	var allUsers []*github.User
	opts := &github.ListOptions{PerPage: 100}

	for {
		users, resp, err := listFunc(gm.ctx, gm.username, opts)
		if err != nil {
			return nil
		}

		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allUsers
}

func (gm *GitHubManager) setCheckedUser(login string, value bool) {
	gm.usersMutex.Lock()
	defer gm.usersMutex.Unlock()
	gm.checkedUsers[login] = value
}

func (gm *GitHubManager) isCheckedUser(login string) bool {
	gm.usersMutex.RLock()
	defer gm.usersMutex.RUnlock()
	return gm.checkedUsers[login]
}

func main() {
	fmt.Print("GitHub 토큰을 입력하세요: ")
	var token string
	fmt.Scan(&token)

	manager, err := NewGitHubManager(token)
	if err != nil {
		fmt.Printf("Error creating GitHub manager: %v\n", err)
		return
	}

	manager.Run()
}
