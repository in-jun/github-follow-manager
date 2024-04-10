# GitHub Follow Manager

이 도구는 GitHub 사용자를 자동으로 팔로우하고 언팔로우하는 기능을 제공합니다.

## 사용법

1. **GitHub 액세스 토큰 생성:**

    - 이 도구를 사용하려면 먼저 [GitHub Personal Access Token](https://docs.github.com/ko/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)을 생성해야 합니다.
    - 생성한 토큰은 프로그램 실행 시 입력해야 합니다.

2. **프로그램 실행:**

    - 터미널 또는 명령 프롬프트에서 프로젝트 디렉터리로 이동한 후, 다음 명령을 실행합니다:
        ```bash
        go run .
        ```
    - 토큰 입력 프롬프트가 나타나면 생성한 GitHub 액세스 토큰을 입력합니다.

3. **자동 팔로우 및 언팔로우 확인:**
    - 프로그램은 주어진 사용자의 팔로워들 중 팔로잉하지 않은 사용자를 찾아 자동으로 팔로우합니다.
    - 또한, 프로그램은 주어진 사용자의 팔로잉들 중 맞팔을 끊은 사용자를 찾아 자동으로 언팔로우합니다.

## 주의사항

-   GitHub의 [사용 약관](https://docs.github.com/ko/site-policy/github-terms/github-terms-of-service)을 준수해야 합니다.
