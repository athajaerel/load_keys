package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
)

// Global vars for modification by build system
var build_mode string
var app_name = "dummy"
var app_exename = "dummy.exe"
var app_version = "dummy_ver"
var conf_default_file = "/etc/dummy.yml"

const docopt_str = `%s.

Usage:
  %s
  %s -c <conf_file>
  %s -h | --help
  %s -v | --version

Options:
  -h --help       Show this screen.
  -v --version    Show version.
  -c <conf_file>  Configuration file [default: %s].
`

const docopt_name = `%s %s`

func main() {
	var loglevel _loglevel
	loglevel = INFO
	if build_mode == "Debug" {
		loglevel = DEBUG
	}
	// TODO: correct this so not repeating var name
	args, _ := docopt.ParseArgs(
		fmt.Sprintf(
			docopt_str,
			app_name,
			app_exename,
			app_exename,
			app_exename,
			app_exename,
			conf_default_file),
		nil,
		fmt.Sprintf(
			docopt_name,
			app_name,
			app_version))
	c := Control{
		args: args,
		loglevel: loglevel}
	c.begin()
	c.run()
	c.end()
}
