package main

import (
	"strings"
	"github.com/docopt/docopt-go"
)

// Global vars for modification by build system
var build_mode string
var app_name = "dummy"
var app_exename = "dummy.exe"
var app_version = "dummy_ver"
var conf_default_file = "/etc/dummy.yml"

const docopt_str = `<app_name>.

Usage:
  <app_exename>
  <app_exename> -c <conf_file>
  <app_exename> -h | --help
  <app_exename> -v | --version

Options:
  -h --help       Show this screen.
  -v --version    Show version.
  -c <conf_file>  Configuration file [default: <conf_default_file>].
`

const docopt_version = `<app_name> <app_version>`

// TODO: implement Pythonic f-strings? eg. f("123{num}789"), or whatever. Is it even possible?

func main() {
	var loglevel _loglevel
	loglevel = INFO
	if build_mode == "Debug" {
		loglevel = DEBUG
	}
	docopt_str_mod := docopt_str
	docopt_str_mod = strings.Replace(docopt_str_mod, "<app_name>", app_name, -1)
	docopt_str_mod = strings.Replace(docopt_str_mod, "<app_exename>", app_exename, -1)
	docopt_str_mod = strings.Replace(docopt_str_mod, "<conf_default_file>", conf_default_file, -1)
	docopt_version_mod := docopt_version
	docopt_version_mod = strings.Replace(docopt_version_mod, "<app_name>", app_name, -1)
	docopt_version_mod = strings.Replace(docopt_version_mod, "<app_version>", app_version, -1)
	args, _ := docopt.ParseArgs(
		docopt_str_mod,
		nil,
		docopt_version_mod)
	c := Control{
		args: (map[string]interface{})(args),
		loglevel: loglevel}
	c.run()
}
