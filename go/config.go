package main

// Custom config for this application.
// Too much effort to make generic library.

import (
	"github.com/FatmanUK/fatgo/serialisers"
)

type Config struct {
	version float32
	secret string
	keys string
	wisp string
	bin_env string
	bin_agent string
	bin_add string
}

const conf_version = 0.2
const conf_secret = "vaults/secret.txt"
const conf_keys = "vaults/keys.yml"
var conf_wisp = "/dev/shm/wisp.bash"
var conf_bin_env = "/usr/bin/env"
var conf_bin_agent = "/usr/bin/ssh-agent"
var conf_bin_add = "/usr/bin/ssh-add"

func (re *Config) serialise(s serialisers.Serialiser) {
	s.IoF(&re.version)
	s.IoS(&re.secret)
	s.IoS(&re.keys)
	s.IoS(&re.wisp)
	s.IoS(&re.bin_env)
	s.IoS(&re.bin_agent)
	s.IoS(&re.bin_add)
}

func (re *Config) load(buf *[]byte) {
	re.serialise(&serialisers.Loader{Array: buf})
}

func (re *Config) save(buf *[]byte) {
	re.serialise(&serialisers.Saver{Array: buf})
}

func (re *Config) size() uint64 {
	var buf_size uint64 = 0
	re.serialise(&serialisers.Sizer{&buf_size})
	return buf_size
}
