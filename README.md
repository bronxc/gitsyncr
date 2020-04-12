## About

I contribute to various different open source projects and have multiple
different forks of these projects, so I needed a simple tool for keeping my
forks up-to-date. Hence I made this simple tool called `gitsyncr` (mandatory
Flickr `-r`)

## Installation

```
go get github.com/topikettunen/gitsyncr
cd $GOPATH/src/github.com/topikettunen/gitsyncr
go build # or go install
```

## Configuration

`gitsyncr` reads config file by default from `~/.gitsyncr` but this can be
configured with the `GITSYNCR_CONFIG` environment variable. Sample
configurations can be found from `sample.toml` and it looks something like this:

``` toml
[user]
key = "~/.ssh/id_rsa"

[forks]
  [forks.linux]
  upstream = "git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git"
  fork = "git@github.com:topikettunen/linux.git"

  [forks.kubernetes]
  upstream = "git@github.com:kubernetes/kubernetes.git"
  fork = "git@github.com:topikettunen/kubernetes.git"
```

At the moment `gitsyncr` only supports SSH authentication since that is
something that I mainly use for my development.

## Running

After build you can run the compiled binary, which then reads your config file
and clones them in to specified directory. By default it clones them to your `HOME`
but it can be modified with `GITSYNCR_FORK_DIR`. If the repo is already existing
in the `GITSYNCR_FORK_DIR` it pull changes to them instead. At the moment
`gitsyncr` pull changes from `upstream` remote and `master` branch, so if your repository doesn't
have that you should add them. If `gitsyncr` clones the repo, it makes this
remote for your.

After repository is cloned, or changes are pulled, it pushes these changes to
`fork` remote's `master` branch. So your fork's master should up-to-date after
the run.

Personally I run this daily, so I know that almost all the times my fork stays
up-to-date.

## TODO

- Assume different remote names

	- At the moment it only uses `upstream` and `fork`
	
- Fork specific location and branches

- Various refactorings
