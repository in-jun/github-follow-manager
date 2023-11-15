package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

const maxRetries = 10

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Print("GitHub 토큰을 입력하세요: ")
	var token string
	fmt.Scan(&token)

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("인증된 사용자 정보를 가져오는 중 오류 발생: %v\n", err)
		return
	}

	followerFollowers, err := getFollowersOfFollowers(ctx, client, *user.Login)
	if err != nil {
		fmt.Printf("팔로우 할 사람을 얻어오는중 오류발생: %v\n", err)
		return
	}

	fmt.Printf("총 %v 명\n", len(followerFollowers))

	for _, follow := range followerFollowers {
		err := followWithExponentialBackoff(ctx, client, *follow.Login)
		if err != nil {
			fmt.Printf("%s를 팔로우하는 중 오류 발생: %v\n", *follow.Login, err)
			continue
		}
		fmt.Printf("- %s를 팔로우했습니다\n", *follow.Login)
		time.Sleep(1 * time.Second) // 1초 대기
	}
}

func getFollowersOfFollowers(ctx context.Context, client *github.Client, username string) ([]*github.User, error) {
	var allFollowersOfFollowers []*github.User

	following, err := getAllFollowing(ctx, client, username)
	if err != nil {
		return nil, err
	}

	followers, err := getAllFollowers(ctx, client, username)
	if err != nil {
		return nil, err
	}

	followingSet := make(map[string]bool)
	for _, user := range following {
		followingSet[*user.Login] = true
	}

	for _, follower := range followers {
		opts := &github.ListOptions{
			Page:    1,
			PerPage: 100,
		}
		followerFollowers, _, err := client.Users.ListFollowers(ctx, *follower.Login, opts)
		if err != nil {
			fmt.Printf("%s의 팔로워 목록을 가져오는 중 오류 발생: %v\n", *follower.Login, err)
			continue
		}

		for _, followerFollower := range followerFollowers {
			if followingSet[*followerFollower.Login] {
				fmt.Printf("- 이미 %s를 팔로우하고 있습니다 (%s의 팔로워)\n", *followerFollower.Login, *follower.Login)
				continue
			}

			followingSet[*followerFollower.Login] = true
			allFollowersOfFollowers = append(allFollowersOfFollowers, followerFollower)
		}
	}

	return allFollowersOfFollowers, nil
}

func getAllFollowers(ctx context.Context, client *github.Client, username string) ([]*github.User, error) {
	var allFollowers []*github.User

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		followers, resp, err := client.Users.ListFollowers(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("팔로워 목록을 가져오는 중 오류 발생: %v", err)
		}

		allFollowers = append(allFollowers, followers...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFollowers, nil
}

func getAllFollowing(ctx context.Context, client *github.Client, username string) ([]*github.User, error) {
	var allFollowing []*github.User

	opts := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}

	for {
		following, resp, err := client.Users.ListFollowing(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("팔로잉 목록을 가져오는 중 오류 발생: %v", err)
		}

		allFollowing = append(allFollowing, following...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFollowing, nil
}

func followWithExponentialBackoff(ctx context.Context, client *github.Client, username string) error {
	for retries := 0; retries < maxRetries; retries++ {
		_, err := client.Users.Follow(ctx, username)
		if err == nil {
			return nil
		} else if IsRateLimitError(err) {
			fmt.Printf("속도 제한 오류 남은 재시도 횟수: %d\n", maxRetries-retries-1)
			time.Sleep(time.Duration(retries+1) * time.Minute)
		} else {
			return fmt.Errorf("%s를 팔로우하는 중 오류 발생: %v", username, err)
		}
	}
	return fmt.Errorf("최대 재시도 횟수에 도달했습니다. %s를 팔로우할 수 없습니다", username)
}

func IsRateLimitError(err error) bool {
	if apiError, ok := err.(*github.ErrorResponse); ok {
		return apiError.Response.StatusCode == 429 || apiError.Response.StatusCode == 403 || apiError.Response.StatusCode == 500
	}
	return false
}
