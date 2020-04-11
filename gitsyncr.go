package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"reflect"
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
	Username string
	Password string
	Key      string
}

// TODO(topi): Fork specific locations and branches
type fork struct {
	Upstream string
	Fork     string
}

func main() {
	log.Println("Syncing all your forks...")
	gitsyncrDir := gitsyncrDir()
	forkDir := forkDir()
	// Using default path for now, should be changeable
	cfg := parseConfig(fmt.Sprintf("%s/config.toml", gitsyncrDir))
	for forkName, fork := range cfg.Forks {
		forkPath := fmt.Sprintf("%s/%s", forkDir, forkName)
		log.Printf("Cloning %s to %s...", fork.Upstream, forkPath)
		cloneRepo(fork.Upstream, forkPath, cfg.User)
		addForkRemote(forkPath, fork.Fork)
		log.Printf("Pushing changes to %s...", fork.Fork)
		pushChanges(forkPath, "fork", cfg.User)
	}
}

func gitsyncrDir() string {
	var gitsyncrDir string
	if value, ok := os.LookupEnv("GITSYNCR_CONFIG_DIR"); ok {
		gitsyncrDir = value
	} else {
		home := userHomeDir()
		gitsyncrDir = fmt.Sprintf("%s/.gitsyncr", home)
	}
	if _, err := os.Stat(gitsyncrDir); os.IsNotExist(err) {
		os.Mkdir(gitsyncrDir, os.ModeDir)
	}
	return gitsyncrDir
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

// Config should be read from ~/.gitsyncr/config.toml by default
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

func cloneRepo(url, path string, user user) {
	// Check for already existing repo
	// if existing -> pull
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Printf("%s already existing, pulling changes...\n", path)
		pullChanges(path, "master")
		return
	}
	var publicKey *ssh.PublicKeys
	sshPath := normalizeSSHKeyPath(user.Key)
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, keyError := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		log.Fatal(keyError)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stop
		cancel()
	}()
	log.Println("To gracefully stop the clone operation, push Crtl-C.")
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
	_, err := git.PlainCloneContext(ctx, path, false, &cloneOpts)
	if err != nil {
		log.Fatal(err)
	}
}

func remoteExists(arrayType interface{}, item interface{}) bool {
	arr := reflect.ValueOf(arrayType)
	if arr.Kind() != reflect.Array {
		log.Fatalf("Invalid data-type\n")
	}
	for i := 0; i < arr.Len(); i++ {
		if arr.Index(i).Interface() == item {
			return true
		}
	}
	return false
}

func addForkRemote(path, url string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	remotes, err := r.Remotes()
	if err != nil {
		log.Fatal(err)
	}
	if remoteExists(remotes, "fork") {
		return
	} else {
		_, err = r.CreateRemote(&gitConfig.RemoteConfig{
			Name: "fork",
			URLs: []string{url},
		})
		if err != nil {
			log.Fatal(err)
		}
		return
	}
}

// TODO(topi): Pulling fails if up to date when pulling from git://
func pullChanges(path, branch string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}
	err = w.Pull(&git.PullOptions{
		RemoteName:    "upstream",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatal(err)
	}
}

// TODO(topi): Pushing hangs without error
func pushChanges(path, remote string, user user) {
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Fatal(err)
	}
	var publicKey *ssh.PublicKeys
	sshPath := normalizeSSHKeyPath(user.Key)
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, keyError := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		log.Fatal(keyError)
	}
	err = r.Push(&git.PushOptions{
		RemoteName: remote,
		Auth: publicKey,
		Progress: os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatal(err)
	}
}
