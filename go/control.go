package main

// TODO: the size of this import bit suggests I have too many responsibilities
// going on in this module. Split off some classes.
import (
	"os"
	"fmt"
	"strings"
	"syscall"
	"path"
	"path/filepath"
	"errors"
//	//"github.com/bitfield/script"
	"golang.org/x/term"
	"io/fs"
	"io/ioutil"
	"strconv"
)

type hashmap map[string]string

type Control struct {
	args map[string]interface{}
	model *Model
	view *View
	loglevel _loglevel
	data *Config
	conf_name string
}

// TODO: think about: maybe this should be a standard part of the Model? (env processing)
// Alternatively, get rid. Might not be worth doing at all.
func get_env_vars() hashmap {
	var env hashmap
	env = make(hashmap)
	// TODO: get all defined user env vars
	env["USER"] = os.Getenv("USER")
// Answer: turns out UID is not an env var, it's a bash internal variable. Goddamnit bash.
//	env["UID"] = os.Getuid() // BUG: This is returning nothing for some reason? Try os.Environ instead?
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
	re.view.log(DEBUG, "UID var: " + os.Getenv("UID"))
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

// TODO: could put these three in a general "file utilities" library module.
func (re *Control) list_files_with_prefix(dir string, prefix string) []string {
	files, err := ioutil.ReadDir(dir)
	die_if(err, re.view)
	var rv []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) {
			rv = append(rv, path.Join(dir, f.Name()))
		}
	}
	return rv
}

func (re *Control) is_socket_file(stat fs.FileInfo) bool {
	file_type_mode := stat.Mode() & fs.ModeSocket
	msg := "is socket: "
	if file_type_mode == fs.ModeSocket {
		msg += "yes"
	} else {
		msg += "no"
	}
	re.view.log(DEBUG, msg)
	return (file_type_mode == fs.ModeSocket)
}

func (re *Control) is_owned(stat fs.FileInfo, want_uid int) bool {
	// Very weird that there's no other way to do this
	s, _ := stat.Sys().(*syscall.Stat_t)
	msg := "is owned by running user: "
	if int(s.Uid) == want_uid {
		msg += "yes"
	} else {
		msg += "no ("
		msg += strconv.FormatInt(int64(s.Uid), 10)
		msg += "/"
		msg += strconv.FormatInt(int64(want_uid), 10)
		msg += ")"
	}
	re.view.log(DEBUG, msg)
	return (int(s.Uid) == want_uid)
}

// TODO: these two can go in a "ssh-agent finder" module. Too specific to be a library.
func (re *Control) find_ssh_agents() []string {
	agent_dirs := re.list_files_with_prefix("/tmp", "ssh-X")
	var my_agents []string
	//env := get_env_vars()
	for _, f := range agent_dirs {
		list := re.list_files_with_prefix(f, "agent.")
		// don't like n^2 funcs, but oh well :(
		// should(tm) be small n at least
		for _, g := range list {
			stat, _ := os.Stat(g)
			if re.is_socket_file(stat) && re.is_owned(stat, os.Getuid()) {
				my_agents = append(my_agents, g)
			}
		}
	}
	return my_agents
}

func (re *Control) get_owned_ssh_agents() []string {
	my_agents := re.find_ssh_agents()
	re.view.log(DEBUG, "There are " + strconv.FormatInt(int64(len(my_agents)), 10) + " agents")
	if len(my_agents) == 0 {
		// fork into agent if none owned
		var argv []string
		argv = append(argv, re.data.bin_env)
		argv = append(argv, re.data.bin_agent)

		// TODO: test this where there are no agents, not sure it works
		p := &syscall.ProcAttr{ }
		syscall.ForkExec(argv[0], argv, p)
		re.view.log(ERROR, "Fork failed.")
		die_if(nil, re.view) // shouldn't make it here --- always die

		my_agents = re.find_ssh_agents()
	}
	if len(my_agents) == 0 {
		re.view.log(ERROR, "Couldn't start agent.")
		die_if(nil, re.view) // always die
	}
	return my_agents
}

// TODO: put in a generic utility library.
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

// This is the main flow function. As such, probably needs to remain here. Or re-integrate into parent. Or dis-integrate into minor functions.
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
	re.view.log(DEBUG, "Agents available:")
	for _, f := range my_agents {
		re.view.log(DEBUG, f)
	}


	// Get secret from file or keyboard.
//	password := re.get_password(secret_file)
        // don't print the password you dolt!
	//re.view.log(DEBUG, "Password: " + (string)(password))


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

// TODO: could possibly put these in Config. Is it worth the effort? Maybe.
func (re *Control) load_config() {
	_, err := os.Stat(re.conf_name)
	// TODO: move defaults to Config module?
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
	re.data.load(&buf)
	re.view.log(
		DEBUG,
		fmt.Sprintf("Loading done"))
}

func (re *Control) save_config() {
	re.view.log(
		DEBUG,
		fmt.Sprintf("Saving config: %s", re.conf_name))
	buf := make([]byte, re.data.size())
	// Should the buffer be the responsibility of the Model object?
	re.data.save(&buf)
	write_binary_file(&buf, re.conf_name, re.view)
}
