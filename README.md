# GitHub Follow Manager

GitHub 사용자의 팔로워 및 팔로잉 관계를 자동으로 관리하는 도구입니다.

## 기능

-   팔로워 중 팔로우하지 않은 사용자 자동 팔로우
-   맞팔로우를 끊은 사용자 자동 언팔로우
-   GitHub API 사용량 제한 자동 관리
-   안전한 토큰 기반 인증

## 시작하기

### 사전 요구사항

-   Go 1.16 이상
-   GitHub 계정
-   [GitHub Personal Access Token](https://docs.github.com/ko/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) (최소 권한: `user:follow`)

### 설치

```bash
git clone https://github.com/in-jun/github-follow-manager.git
cd github-follow-manager
```

### 실행

```bash
go run .
```

프롬프트가 표시되면 GitHub Personal Access Token을 입력하세요.

## 면책 조항

이 도구를 사용할 때는 [GitHub 이용약관](https://docs.github.com/ko/site-policy/github-terms/github-terms-of-service)을 준수해야 합니다. 과도한 API 요청이나 자동화된 작업으로 인한 계정 제한은 사용자의 책임입니다.
