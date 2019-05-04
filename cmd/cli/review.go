package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	sw "github.com/sah4ez/go-bitbucket"
	"github.com/urfave/cli"
)

func Review(client *sw.APIClient) cli.Command {
	return cli.Command{
		Name:        "review",
		Aliases:     []string{"r"},
		Description: "start review",
		Action: func(c *cli.Context) error {
			prs, _, err := client.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsGet(Auth, Company, Repo, map[string]interface{}{"state": "OPEN"})
			if err != nil {
				return err
			}

			if LeftBranch == "" || RightBranch == "" {
				return fmt.Errorf("missing LEFT_BRANCH or RIGHT_BRANCH env varibale for compare. LEFT_BRANCH default master")
			}

			for _, pr := range prs.Values {
				if RightBranch == pr.Source.Branch.Name && LeftBranch == pr.Destination.Branch.Name {
					files, err := LoadFiles()
					if err != nil {
						return err
					}
					for _, file := range files {
						fmt.Println("id", pr.Id)
						Diff(file)
					}
				}
			}
			return nil
		},
	}
}

func Diff(file string) {
	os.Setenv(EnvCurrentFilenameDiff, file)
	defer os.Unsetenv(EnvCurrentFilenameDiff)

	cmd := exec.Command("git", "difftool", "--tool=vimdiff2", "origin/"+LeftBranch, "origin/"+RightBranch, file)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	r := bufio.NewReader(os.Stdout)

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	line, _, err := r.ReadLine()
	for err != nil {
		fmt.Println(line)
		text, _ := reader.ReadString('\r')
		os.Stdin.Write([]byte(text + "\n"))
		line, _, err = r.ReadLine()
	}

	cmd.Wait()
}
