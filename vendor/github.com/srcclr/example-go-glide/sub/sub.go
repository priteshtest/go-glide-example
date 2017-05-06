package sub

import "strconv"
import "fmt"
import "github.com/google/go-github/github"

func Foo() {
    client := github.NewClient(nil)

    // list public repositories for org "github"
    opt := &github.RepositoryListByOrgOptions{Type: "public"}
    repos, _, _ := client.Repositories.ListByOrg("github", opt)
    
    fmt.Println("Hello from sub! Num repos: " + strconv.Itoa(len(repos)))
}
