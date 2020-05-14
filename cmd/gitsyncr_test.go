package main

import "testing"

func TestParseConfig(t *testing.T) {
	config := parseConfig("sample.toml")
	tests := []struct {
		in, want string
	}{
		{config.User.Key, "~/.ssh/id_rsa"},
		{config.Forks["linux"].Upstream, "git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git"},
		{config.Forks["linux"].Fork, "git@github.com:topikettunen/linux.git"},
		{config.Forks["kubernetes"].Upstream, "git@github.com:kubernetes/kubernetes.git"},
		{config.Forks["kubernetes"].Fork, "git@github.com:topikettunen/kubernetes.git"},
	}
	for _, tt := range tests {
		if tt.in != tt.want {
			t.Errorf("Got %q, wanted %q", tt.in, tt.want)
		}
	}
}
