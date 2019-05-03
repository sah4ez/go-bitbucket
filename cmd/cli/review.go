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

			if c.Args().Get(0) == "" {
				return fmt.Errorf("missing repo branch which would compare with master")
			}

			for _, pr := range prs.Values {
				if c.Args().Get(0) == pr.Source.Branch.Name {
					cmd := exec.Command("git", "difftool", "--tool=vimdiff2", "origin/"+pr.Destination.Branch.Name, "origin/"+pr.Source.Branch.Name)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					r := bufio.NewReader(os.Stdout)

					err = cmd.Start()
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println("id", pr.Id)

					reader := bufio.NewReader(os.Stdin)

					line, _, err := r.ReadLine()
					for err != nil {
						fmt.Println(line)
						text, _ := reader.ReadString('\n')
						os.Stdin.Write([]byte(text + "\n"))
						line, _, err = r.ReadLine()
					}

					cmd.Wait()
				}
			}
			return nil
		},
	}
}
