package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
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
	prID, _ := strconv.ParseInt(os.Getenv("PR"), 10, 32)

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

	pager, resp, err := client.PullrequestsApi.RepositoriesUsernameRepoSlugPullrequestsPullRequestIdCommentsGet(auth, "ronte", repo, int32(prID))

	if err != nil {
		panic(err.Error())
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(b))
	files := map[string]int{}
	commentsTo := map[string][]string{}
	commentsFrom := map[string][]string{}

	for _, comment := range pager.Values {
		if _, ok := files[comment.Inline.Path]; !ok {
			f, err := os.Open(comment.Inline.Path)
			if err != nil {
				fmt.Println("err", err)
				continue
			}
			lines, err := lineCounter(f)
			if err != nil {
				fmt.Println("err", err)
				continue
			}
			files[comment.Inline.Path] = lines
			commentsTo[comment.Inline.Path] = make([]string, lines)
			commentsFrom[comment.Inline.Path] = make([]string, lines)
		}

		if comment.Inline.To > 0 {
			c := "(" + comment.User.Username + ")" +
				comment.CreatedOn.Format("2006-01-02/15:04") +
				":" + fmt.Sprintf("%d", comment.Id) +
				"  " + comment.Content.Raw
			commentsTo[comment.Inline.Path][comment.Inline.To-1] += " << " + c
		}

		if comment.Inline.From > 0 {
			c := "(" + comment.User.Username + ")" +
				comment.CreatedOn.Format("2006-01-02/15:04") +
				":" + fmt.Sprintf("%d", comment.Id) +
				"  " + comment.Content.Raw
			commentsFrom[comment.Inline.Path][comment.Inline.From-1] += " >> " + c
		}

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

	for name, comment := range commentsTo {
		data := []byte(strings.Join(comment, "\n"))
		parts := strings.Split(name, "/")
		err := ioutil.WriteFile("/tmp/to_"+parts[len(parts)-1]+".comments", data, 0644)
		if err != nil {
			fmt.Println("err", err)
			continue
		}
	}

	for name, comment := range commentsFrom {
		data := []byte(strings.Join(comment, "\n"))
		parts := strings.Split(name, "/")
		err := ioutil.WriteFile("/tmp/from_"+parts[len(parts)-1]+".comments", data, 0644)
		if err != nil {
			fmt.Println("err", err)
			continue
		}
	}
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
