package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	sw "github.com/sah4ez/go-bitbucket"
)

func main() {
	cred := os.Getenv("BITBUCKET_CREDENTIAL")
	parts := strings.Split(cred, ":")
	auth := context.WithValue(context.Background(), sw.ContextBasicAuth, sw.BasicAuth{
		UserName: parts[0],
		Password: parts[1],
	})

	repo := os.Getenv("REPO")
	config := sw.NewConfiguration()

	client := sw.NewAPIClient(config)

	if len(os.Args) == 3 && os.Args[1] == "review" {
		prs, _, err := client.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsGet(auth, "ronte", repo, map[string]interface{}{"state": "OPEN"})
		if err != nil {
			panic(err.Error())
		}
		for _, pr := range prs.Values {
			if os.Args[2] == pr.Source.Branch.Name {
				cmd := exec.Command("git", "difftool", "--tool=vimdiff2", "origin/"+pr.Source.Branch.Name, "origin/"+pr.Destination.Branch.Name)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				r := bufio.NewReader(os.Stdout)

				// Start the command!
				err = cmd.Start()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("id", pr.Id)

				line, _, err := r.ReadLine()
				reader := bufio.NewReader(os.Stdin)

				for err != nil {
					fmt.Println(line)
					text, _ := reader.ReadString('\n')
					os.Stdin.Write([]byte(text + "\n"))
					line, _, err = r.ReadLine()
				}

				cmd.Wait()
			}
		}
		os.Exit(0)
	}

	pager, resp, err := client.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsPullRequestIdCommentsGet(auth, "ronte", repo, 62)

	if err != nil {
		panic(err.Error())
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(b))
	for _, comment := range pager.Values {
		fmt.Println("raw", comment.Content.Raw)
		if comment.Parent != nil {
			fmt.Println("id", comment.Parent.Id)
		}
		fmt.Println("username", comment.User.Username)
		fmt.Println("to", comment.Inline.To)
		fmt.Println("from", comment.Inline.From)
	}
}
