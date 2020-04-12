package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type config struct {
	User  user
	Forks map[string]fork
}

type user struct {
	Key      string
}

// TODO(topi): Fork specific locations and branches
type fork struct {
	Upstream string
	Fork     string
}

func main() {
	log.Println("Syncing all your forks...")
	gitsyncrConfig := gitsyncrConfig()
	forkDir := forkDir()
	cfg := parseConfig(gitsyncrConfig)
	publicKey := newPublicKeys(cfg.User.Key)
	for forkName, fork := range cfg.Forks {
		forkPath := fmt.Sprintf("%s/%s", forkDir, forkName)
		// Check for already existing repo
		// if existing -> pull
		if _, err := os.Stat(forkPath); !os.IsNotExist(err) {
			log.Printf("%s already existing, pulling changes...\n", forkPath)
			checkRemote(forkPath, fork.Upstream, "upstream")
			checkRemote(forkPath, fork.Fork, "fork")
			pullChanges(fork.Upstream, forkPath, "master", cfg.User, publicKey)
		} else {
			log.Printf("Cloning %s to %s...", fork.Upstream, forkPath)
			cloneRepo(fork.Upstream, forkPath, cfg.User, publicKey)
			checkRemote(forkPath, fork.Upstream, "upstream")
			checkRemote(forkPath, fork.Fork, "fork")
		}
		log.Printf("Pushing changes to %s...", fork.Fork)
		pushChanges(fork.Fork, forkPath, cfg.User, publicKey)
	}
}

func gitsyncrConfig() string {
	var gitsyncrConfig string
	if value, ok := os.LookupEnv("GITSYNCR_CONFIG"); ok {
		gitsyncrConfig = value
	} else {
		home := userHomeDir()
		gitsyncrConfig = fmt.Sprintf("%s/.gitsyncr", home)
	}
	return gitsyncrConfig
}

func forkDir() string {
	if value, ok := os.LookupEnv("GITSYNCR_FORK_DIR"); ok {
		return value
	} else {
		return userHomeDir()
	}
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return home
}

// Config should be read from ~/.gitsyncr by default
func parseConfig(cfgPath string) config {
	var config config
	if _, err := toml.DecodeFile(cfgPath, &config); err != nil {
		log.Fatal(err)
	}
	return config
}

func normalizeSSHKeyPath(path string) string {
	if strings.Contains(path, "~") {
		home := userHomeDir()
		// if path starts with // it should go to root, but just to be safe
		if home == "/" {
			path = strings.Replace(path, "", home, 1)
			return path
		}
		path = strings.Replace(path, "~", home, 1)
		return path
	} else {
		return path
	}
}

func newPublicKeys(key string) *ssh.PublicKeys {
	var publicKeys *ssh.PublicKeys
	sshPath := normalizeSSHKeyPath(key)
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKeys, keyError := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		log.Fatal(keyError)
	}
	return publicKeys
}	

func cloneOpts(url string, user user, publicKey *ssh.PublicKeys) git.CloneOptions {
	var cloneOpts git.CloneOptions
	// This could probably be done cleaner.
	// Upstream fork should be cloned with a upstream remote name for easier
	// distinction between remotes (personal preference).
	if strings.Contains(url, "git://") {
		cloneOpts = git.CloneOptions{
			URL:           url,
			Progress:      os.Stdout,
			RemoteName:    "upstream",
			ReferenceName: plumbing.NewBranchReferenceName("master"),
			SingleBranch:  true,
		}
	} else {
		cloneOpts = git.CloneOptions{
			Auth:          publicKey,
			URL:           url,
			Progress:      os.Stdout,
			RemoteName:    "upstream",
			ReferenceName: plumbing.NewBranchReferenceName("master"),
			SingleBranch:  true,
		}
	}
	return cloneOpts
}

func cloneRepo(url, path string, user user, publicKey *ssh.PublicKeys) {
	cloneOpts := cloneOpts(url, user, publicKey)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stop
		cancel()
	}()
	log.Println("To gracefully stop the clone operation, push Crtl-C.")
	_, err := git.PlainCloneContext(ctx, path, false, &cloneOpts)
	if err != nil {
		log.Fatal(err)
	}
}

func checkRemote(path, url, remote string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating %s remote to %s...\n", remote, path)
	_, err = r.CreateRemote(&gitConfig.RemoteConfig{
		Name: remote,
		URLs: []string{url},
	})
	if err == git.ErrRemoteExists {
		log.Printf("%s remote already exists in %s, continuing...\n", remote, path)
	} else if err != nil {
		log.Fatal(err)
	}
}

func pullOpts(url, branch string, user user, publicKey *ssh.PublicKeys) git.PullOptions {
	var pullOpts git.PullOptions
	// This could probably be done cleaner.
	// Upstream fork should be cloned with a upstream remote name for easier
	// distinction between remotes (personal preference).
	if strings.Contains(url, "git://") {
		pullOpts = git.PullOptions{
			RemoteName:    "upstream",
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			Progress: os.Stdout,
		}
	} else {
		pullOpts = git.PullOptions{
			RemoteName: "upstream",
			Auth: publicKey,
			Progress: os.Stdout,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
		}
	}
	return pullOpts
}

func pullChanges(url, path, branch string, user user, publicKey *ssh.PublicKeys) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	pullOpts := pullOpts(url, branch, user, publicKey)
	err = w.Pull(&pullOpts)
	if err == git.NoErrAlreadyUpToDate {
		log.Printf("%s already up to date...\n", path)
	} else if err != nil {
		log.Fatal(err)
	}
}

func pushOpts(url string, user user, publicKey *ssh.PublicKeys) git.PushOptions{
	var pushOpts git.PushOptions
	// This could probably be done cleaner.
	// Upstream fork should be cloned with a upstream remote name for easier
	// distinction between remotes (personal preference).
	if strings.Contains(url, "git://") {
		pushOpts = git.PushOptions{
			RemoteName: "fork",
			Progress: os.Stdout,
		}
	} else {
		pushOpts = git.PushOptions{
			RemoteName: "fork",
			Auth: publicKey,
			Progress: os.Stdout,
		}
	}
	return pushOpts
}

func pushChanges(url, path string, user user, publicKey *ssh.PublicKeys) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	pushOpts := pushOpts(url, user, publicKey)
	err = r.Push(&pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		log.Printf("%s already up to date...\n", url)
	} else if err != nil {
		log.Fatal(err)
	}
}
