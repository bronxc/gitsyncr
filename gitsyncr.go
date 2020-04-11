package main

import (
	"os"
	"log"
	"github.com/go-git/go-git/v5" 
)

func cloneRepo(url string, path string) {
	log.Printf("Cloning %s to %s...", url, path)
	// Check for already existing repo
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Println("Syncing all your forks...")
	cloneRepo("https://github.com/topikettunen/brutal-emacs", "brutal-emacs")
}
