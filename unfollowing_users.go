package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/go-github/v56/github"
)

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
