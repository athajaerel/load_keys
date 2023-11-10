package main

import (
	"os"
	"path"
	"path/filepath"
	"errors"
	"strings"
	"fmt"
	"github.com/docopt/docopt-go"
	//"github.com/bitfield/script"
	"github.com/FatmanUK/fatgo/serialisers"
	"syscall"
	"golang.org/x/term"
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

func (re *Control) run() {
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
	re.view.log(INFO, "Starting program.")
	re.load_keys()
	re.view.log(INFO, "Stopping program.")
}

func (re *Control) find_ssh_agents() []string {
/*
  # find all ssh agents
  agent_dirs=glob(r'%s/ssh-*' % TMPDIR)
  my_agents=[]
  for agent_dir in agent_dirs:
    # must be a directory not a symlink
    if not isdir(agent_dir):
      continue
    # must be owned by user
    if not file_owner(agent_dir) == USER:
      continue
    # must have agent.* file
    agent_files=glob(r'%s/agent.*' % (agent_dir))
    for agent_file in agent_files:
      # must be owned, and a socket file
      if not issock(agent_file):
        continue
      if not file_owner(agent_file) == USER:
        continue
      my_agents+=[agent_file]
  return my_agents
*/
	// BUG hardcode until implemented, will need updating each boot so do it soon :(
	arr := []string{
		"/tmp/ssh-XXXXXX091x9C/agent.3410322",
		"/tmp/ssh-XXXXXXTBJjny/agent.3410291",
		"/tmp/ssh-XXXXXXrtcUhf/agent.6954",
		}
	return arr[:]
}

func (re *Control) get_owned_ssh_agents() []string {
	my_agents := re.find_ssh_agents()
	if len(my_agents) == 0 {
		// fork into agent if none owned
/*
  rside, wside = pipe()
  if not fork():
    childpostfork(rside, wside);
    # Execute the desired program, replace the program image,
    # doesn't return
    execve(BIN_ENV, [BIN_ENV, BIN_AGENT], environ)
    raise ValueError('Failed to exec ssh-agent')
  parentpostfork(rside, wside)
*/
		my_agents = re.find_ssh_agents()
	}
	if len(my_agents) == 0 {
		re.view.log(ERROR, "Couldn't start agent.")
		die_if(nil, re.view) // always die
	}
	return my_agents
}

func (re *Control) get_password(secret_file string) string {
	password, err := os.ReadFile(secret_file)
	if err != nil {
		re.view.log(INFO, "Vault password: ")
		password, err = term.ReadPassword(syscall.Stdin)
	}
	if string(password) == "" {
		re.view.log(ERROR, "Empty password entered.")
		die_if(err, re.view)
	}
	return string(password)
}

func (re *Control) load_keys() {
	me_dir := filepath.Dir(os.Args[0])
	re.view.log(DEBUG, "Path: " + me_dir)
	//secret_file := me_dir + "/" + re.data.secret
	//keys_file := me_dir + "/" + re.data.keys
	// Bug in serialiser, loads the string twice?
	// Hardcode until fixed.
	secret_file := path.Join(me_dir, "vaults/secret.txt")
	keys_file := path.Join(me_dir, "vaults/keys.yml")
	re.view.log(DEBUG, fmt.Sprintf("Secret: %s", secret_file))
	re.view.log(DEBUG, "Keys: " + keys_file)
	//wisp_path := "/dev/shm/wisp.bash"
	//bin_env := "/usr/bin/env"
	//bin_agent := "/usr/bin/ssh-agent"
	//bin_add := "/usr/bin/ssh-add"

	// Get owned SSH agents, or start one.
	my_agents := re.get_owned_ssh_agents()

	// Get secret from file or keyboard.
	password := re.get_password(secret_file)
	re.view.log(DEBUG, "Password: " + (string)(password))
/*
# get fully decrypted JSON object from vault
vaultblob=slurp(KEYS)
assert vaultblob!=r'', r'Empty vault blob read.'
debug(vaultblob, 'vault blob')

myobj=None
# if it's a JSON file, load it as JSON and decrypt the passwords
if isyaml(vaultblob):
  myobj=objectfromyaml(vaultblob)
  myobj=jsonwalk(myobj, decryptfield, password)
else:
  # otherwise, decrypt as a blob and confirm it's JSON
  # TODO
  vault=Vault(password)
  myobj=vault.load(vaultblob)

assert myobj!=None, r'Nothing returned from vault decryption'

# TODO: assert mounted location is executable

for key in myobj[0]['keys']:
  debug(key, 'Key')
  debug(my_agents, r'ssh agents')
  # expand key['path'], ssh-add doesn't like tildes
  key['path']=expanduser(key['path'])
  for agent in my_agents:
    # create wisp script
    with open(WISP_PATH, r'w') as f:
      f.write((r"""#!%s bash
echo '%s'
/bin/rm %s""" % (BIN_ENV, key['password'], WISP_PATH)))
      f.close()
    chmod(WISP_PATH, 0o755)
    # add key to agent
    rside, wside = pipe()
    # add env vars to environment, so we can use the SSH_ASKPASS trick
    environ.update({ 'SSH_AUTH_SOCK': agent })
    environ.update({ 'SSH_ASKPASS': WISP_PATH })
    environ.update({ 'SSH_ASKPASS_REQUIRE': 'force' })
    environ.update({ 'DISPLAY': '' })
    if not fork():
      childpostfork(rside, wside);
      # Execute the desired program, replace the program image,
      # doesn't return
      execve(BIN_ADD, [BIN_ADD, key['path']], environ)
      raise ValueError('Failed to exec ssh-agent')
    parentpostfork(rside, wside)
*/
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
	// Should the buffer be the responsibility of the Model object?
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
	// Should the buffer be the responsibility of the Model object?
	re.serialise(&serialisers.Saver{Array: &buf})
	write_binary_file(&buf, re.conf_name, re.view)
}
