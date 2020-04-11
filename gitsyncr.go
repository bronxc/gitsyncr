package main

import (
	"os"
	"log"
	
	"github.com/go-git/go-git/v5"
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
	parseConfig("sample.toml")
}

func parseConfig(cfgPath string) config {
	var config config
	if _ , err := toml.DecodeFile(cfgPath, &config); err != nil {
		panic(err)
	}
	return config
}
	

func cloneRepo(url string, path string) {
	log.Printf("Cloning %s to %s...", url, path)
	// Check for already existing repo
	// if existing -> pull
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
		Progress: os.Stdout,
	})
	if err != nil {
		panic(err)
	}
}

func pullChanges() {} 		// What this takes in?
