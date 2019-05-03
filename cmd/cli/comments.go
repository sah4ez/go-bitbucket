package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	sw "github.com/sah4ez/go-bitbucket"
	"github.com/urfave/cli"
)

func Comments(client *sw.APIClient) cli.Command {
	return cli.Command{
		Name:        "comment",
		Aliases:     []string{"c"},
		Description: "load all comments in PR",
		Action: func(c *cli.Context) error {
			pager, resp, err := client.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsPullRequestIdCommentsGet(Auth, Company, Repo, int32(PrID))

			if err != nil {
				return err
			}
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			files := map[string]int{}
			commentsTo := map[string][]string{}
			commentsFrom := map[string][]string{}

			for _, comment := range pager.Values {
				if _, ok := files[comment.Inline.Path]; !ok {
					f, err := os.Open(comment.Inline.Path)
					if err != nil {
						return err
					}
					lines, err := lineCounter(f)
					if err != nil {
						return err
					}
					files[comment.Inline.Path] = lines
					commentsTo[comment.Inline.Path] = make([]string, lines)
					commentsFrom[comment.Inline.Path] = make([]string, lines)
				}

				if comment.Inline.To > 0 {
					commentsTo[comment.Inline.Path][comment.Inline.To-1] += BuildCommentLine("<", comment)
				}

				if comment.Inline.From > 0 {
					commentsFrom[comment.Inline.Path][comment.Inline.From-1] += BuildCommentLine(">", comment)
				}
				PrintDebug(c.GlobalBool("debug"), comment)

			}

			saveComments(c.GlobalBool("debug"), c.GlobalString("prefix"), "to", commentsTo)
			saveComments(c.GlobalBool("debug"), c.GlobalString("prefix"), "from", commentsFrom)

			return nil
		},
	}
}

func saveComments(debug bool, prefix, dest string, comments map[string][]string) error {
	for name, comment := range comments {
		data := []byte(strings.Join(comment, "\n"))
		parts := strings.Split(name, "/")
		fileName := prefix + "/" + dest + "_" + parts[len(parts)-1] + ".comments"
		if debug {
			fmt.Println("save comments to: ", fileName)
		}
		err := ioutil.WriteFile(fileName, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func PrintDebug(need bool, comment sw.PullrequestComment) {
	if need {
		fmt.Println("raw", comment.Content.Raw)
		if comment.Parent != nil {
			fmt.Println("id", comment.Parent.Id)
		}
		fmt.Println("username", comment.User.Username)
		fmt.Println("to", comment.Inline.To)
		fmt.Println("from", comment.Inline.From)
		fmt.Println("path", comment.Inline.Path)
		fmt.Println()
	}
}

func BuildCommentLine(delimiter string, comment sw.PullrequestComment) string {
	// [123] < (sah4ez) < 2019-01-31T01:01:01.0Z < Comment text raw <<<
	return fmt.Sprintf("[%d] %s (%s) %s %s %s %s %s ",
		comment.Id,
		delimiter,
		comment.User.Username,
		delimiter,
		comment.CreatedOn.Format(time.RFC3339),
		delimiter,
		comment.Content.Raw,
		strings.Repeat(delimiter, 3),
	)
}

func lineCounter(r io.Reader) (int, error) {

	var readSize int
	var err error
	var count int

	buf := make([]byte, 1024)

	for {
		readSize, err = r.Read(buf)
		if err != nil {
			break
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], '\n')
			if i == -1 || readSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
	}
	if readSize > 0 && count == 0 || count > 0 {
		count++
	}
	if err == io.EOF {
		return count, nil
	}

	return count, err
}
