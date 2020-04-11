package main

import (
	"os"
	"log"
	"fmt"
	
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/BurntSushi/toml"
)

type config struct {
	User user
	Forks map[string]fork
}

type user struct {
	Username string
	Password string
	Key string
}

type fork struct {
	Upstream string
	Fork string
}

func main() {
	log.Println("Syncing all your forks...")
	// cloneRepo("https://github.com/topikettunen/brutal-emacs", "brutal-emacs")
	cfg := parseConfig("sample.toml")
	for forkName, fork := range cfg.Forks {
		// Get the wanted destination from somewhere
		log.Printf("Cloning %s to ~/.gitsyncr/%s...", fork.Upstream, forkName)
		// ~/.gitsyncr should be made if not existing
		cloneRepo(fork.Upstream, fmt.Sprintf("~/.gitsyncr/%s", forkName))
	}
}

// Config should be read from ~/.gitsyncr/config.toml by default
func parseConfig(cfgPath string) config {
	var config config
	if _ , err := toml.DecodeFile(cfgPath, &config); err != nil {
		log.Fatal(err)
	}
	return config
}

// Forks are cloned to ~/.gitsyncr by default
// TODO(topi): SSH auth fails 
func cloneRepo(url string, path string) {
	// Check for already existing repo
	// if existing -> pull
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		pullChanges(path, "master")
		return
	}
	// Upstream fork should be cloned with a upstream remote name 
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
		Progress: os.Stdout,
		RemoteName: "upstream",
		ReferenceName: plumbing.NewBranchReferenceName("master"),
		SingleBranch: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func addForkRemote(path string, url string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	_, err = r.CreateRemote(&gitConfig.RemoteConfig{
		Name: "fork",
		URLs: []string{url},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func pullChanges(path string, branch string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	err = w.Pull(&git.PullOptions{
		RemoteName: "upstream",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		log.Fatal(err)
	}
}
