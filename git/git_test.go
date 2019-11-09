package git

import (
	"testing"
)

func TestGetOrigin(t *testing.T) {
	gitFile := []byte(`
[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = git@gitlab.com:diamondburned/meistercli.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "master"]
	remote = origin
	merge = refs/heads/master
`)

	u, err := getOrigin(gitFile)
	if err != nil {
		t.Fatal(err)
	}

	if u != "diamondburned/meistercli" {
		t.Fatal("Unexpected origin URL "+u, ", expected diamondburned/meistercli")
	}
}
