package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	sw "github.com/sah4ez/go-bitbucket"
	"github.com/urfave/cli"
)

func CommentLine(client *sw.APIClient) cli.Command {
	return cli.Command{
		Name:        "comment-line",
		Aliases:     []string{"cl"},
		Description: "make comment on line in the file",
		Usage:       "<line> <comment_raw>",
		Action: func(c *cli.Context) error {
			line := c.Args().Get(0)
			path := os.Getenv(EnvCurrentFilenameDiff)
			comment := strings.Join(c.Args()[0:], " ")

			if path == "" {
				return errors.New("You try comment line in direct. First need run 'comments review' command.")
			}

			cmd := exec.Command("git", "config", "--global", "user.name")
			authorB, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}
			author := string(bytes.Replace(authorB, []byte("\n"), []byte(""), -1))

			if LeftBranch == "" || RightBranch == "" {
				return fmt.Errorf("missing LEFT_BRANCH or RIGHT_BRANCH env varibale for compare. LEFT_BRANCH default master")
			}
			files, err := LoadFiles()
			if err != nil {
				return err
			}

			if c.GlobalBool("debug") {
				for _, file := range files {
					fmt.Println(file)
				}
				fmt.Printf("[%s] (%s) %s %s", line, author, path, comment)
			}

			// TODO(sah4ez) 04-05-2019: need add call API method

			return nil
		},
	}
}
