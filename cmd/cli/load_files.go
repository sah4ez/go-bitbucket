package main

import (
	"bytes"
	"os/exec"
	"strings"
)

func LoadFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "origin/"+LeftBranch, "origin/"+RightBranch)
	files, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	files = bytes.Replace(files, []byte("\n\n"), []byte("\n"), -1)
	return strings.Split(string(files), "\n"), nil
}
