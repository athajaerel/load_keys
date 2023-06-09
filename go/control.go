package main

import (
	"os"
	"errors"
	"strings"
	"fmt"
	"github.com/docopt/docopt-go"
	//"github.com/bitfield/script"
	"github.com/athajaerel/load_keys/dreamtrack/serialisers"
)

const conf_version = 0.2
const conf_secret = "vaults/secret.txt"
const conf_keys = "vaults/keys.yml"
var conf_wisp = "/dev/shm/wisp.bash"
var conf_bin_env = "/usr/bin/env"
var conf_bin_agent = "/usr/bin/ssh-agent"
var conf_bin_add = "/usr/bin/ssh-add"

type Config struct {
	version float32
	secret string
	keys string
	wisp string
	bin_env string
	bin_agent string
	bin_add string
}

type Control struct {
	args docopt.Opts
	model *Model
	view *View
	loglevel _loglevel
	data *Config
	conf_name string
}

type hashmap map[string]string

func (re *Control) serialise(s serialisers.Serialiser) {
	s.IoF(&re.data.version)
	s.IoS(&re.data.secret)
	s.IoS(&re.data.keys)
	s.IoS(&re.data.wisp)
	s.IoS(&re.data.bin_env)
	s.IoS(&re.data.bin_agent)
	s.IoS(&re.data.bin_add)
}

func get_env_vars() hashmap {
	var env hashmap
	env = make(hashmap)
	env["USER"] = os.Getenv("USER")
	return env
}

func (re *Control) begin() {
	re.view = &View{
		loglevel: re.loglevel,
		HasTime: false,
		HasPrefix: true}
	re.model = &Model{
		view: re.view}
	env := get_env_vars()
	re.conf_name = re.args["-c"].(string)
	re.conf_name = strings.Replace(re.conf_name, "~/", "/home/" + env["USER"] + "/", 1)
	re.conf_name = strings.Replace(re.conf_name, "~", "/home/", 1)
	re.conf_name = strings.Replace(re.conf_name, "$USER", env["USER"], 1)
	re.load_config()
	re.view.log(INFO, fmt.Sprintf("Config format version: %.1f", re.data.version))
}

func (re *Control) run() {
	re.view.log(INFO, "Starting program.")
	//
	re.view.log(INFO, "Stopping program.")
}

func (re *Control) end() {
}

func (re *Control) load_config() {
	_, err := os.Stat(re.conf_name)
	re.data = &Config{
		conf_version,
		conf_secret,
		conf_keys,
		conf_wisp,
		conf_bin_env,
		conf_bin_agent,
		conf_bin_add}
	if errors.Is(err, os.ErrNotExist) {
		re.save_config()
	}
	re.view.log(
		DEBUG,
		fmt.Sprintf("Loading config: %s", re.conf_name))
	var buf_size uint64 = size_binary_file(re.conf_name, re.view)
	re.view.log(
		DEBUG,
		fmt.Sprintf("Sizing: %d", buf_size))
	buf := make([]byte, buf_size)
	zero_buffer(&buf)
	re.view.log(
		DEBUG,
		fmt.Sprintf("Buffer zeroed"))
	read_binary_file(&buf, re.conf_name, re.view)
	re.view.log(
		DEBUG,
		fmt.Sprintf("Buffer read"))
	re.serialise(&serialisers.Loader{Array: &buf})
	re.view.log(
		DEBUG,
		fmt.Sprintf("Loading done"))
}

func (re *Control) save_config() {
	re.view.log(
		DEBUG,
		fmt.Sprintf("Saving config: %s", re.conf_name))
	var buf_size uint64 = 0
	re.serialise(&serialisers.Sizer{&buf_size})
	buf := make([]byte, buf_size)
	re.serialise(&serialisers.Saver{Array: &buf})
	write_binary_file(&buf, re.conf_name, re.view)
}
