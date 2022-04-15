#!/usr/bin/env python3

from sys import path
from os import environ, stat, chmod, pipe, fork, close, fdopen, dup2
from os import open as osopen, waitpid, O_RDONLY, execve
from os.path import realpath, isdir, exists, isfile, expanduser
from glob import glob
from pwd import getpwuid
from stat import S_ISSOCK
from yaml import load, YAMLError, BaseLoader # requires pyyaml
from ansible.constants import DEFAULT_VAULT_ID_MATCH
from ansible.parsing.vault import VaultSecret, VaultLib
from getpass import getpass

from config import *

USER=environ.get(r'USER')
ME_DIR=path[0]
SECRET=realpath(r'%s/vaults/secret.txt' % ME_DIR)
WISP_PATH=r'/dev/shm/wisp.bash'
BIN_ENV=r'/usr/bin/env'
BIN_AGENT=r'/usr/bin/ssh-agent'
BIN_ADD=r'/usr/bin/ssh-add'

def debug(line, prefix=None):
  if (DEBUG_MODE):
    if (prefix):
      print(r'%s: %s' % (prefix, line))
    else:
      print(line)

def file_owner(filename):
  return getpwuid(stat(filename).st_uid).pw_name

def issock(filename):
  return S_ISSOCK(stat(filename).st_mode)

def slurp(filename):
  blob=r''
  with open(filename, r'rb') as f:
    # slurp up text blob, also strip newline from end
    blob=f.read().decode('utf-8').strip()
    f.close()
  return blob

def loadyaml(myblob):
  return load(myblob, Loader=BaseLoader)

def isyaml(myblob):
  try:
    loadyaml(myblob)
  except YAMLError as exc:
    print(exc)
    return False
  return True

def objectfromyaml(myblob):
  assert isyaml(myblob)==True, r'Not well-formed YAML'
  myyaml=loadyaml(myblob)
  return myyaml

def decryptfield(node, password):
  # decrypt all strings beginning $ANSIBLE_VAULT;1.1;AES256
  splitnode=node.split('\n', 1)
  if (splitnode[0]=='$ANSIBLE_VAULT;1.1;AES256'):
    a = [(DEFAULT_VAULT_ID_MATCH, VaultSecret(password.encode()))]
    v = VaultLib(a)
    v.cipher_name = 'AES256'
    node = v.decrypt(node).decode('utf-8').strip()
  return node

def jsonwalk(node, fn, password):
  if type(node) is dict:
    return {k: jsonwalk(v, fn, password) for k, v in node.items()}
  elif type(node) is list:
    return [jsonwalk(x, fn, password) for x in node]
  else:
    return fn(node, password)

def find_ssh_agents():
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

def parentpostfork(rside, wside):
  close(wside)
  for line in fdopen(rside):
    print(line.strip())
  # reap zombie
  pid, status=waitpid(-1, 0)
  print('Child exited: pid %d returned %d' % (pid, status))

def childpostfork(rside, wside):
  close(rside)
  dup2(wside, 1) # Redirect stdout to parent
  dup2(wside, 2) # Redirect stderr to parent
  devnull=osopen('/dev/null', O_RDONLY)
  dup2(devnull, 0)

my_agents=find_ssh_agents()
if my_agents==[]:
  rside, wside = pipe()
  if not fork():
    childpostfork(rside, wside);
    # Execute the desired program, replace the program image,
    # doesn't return
    execve(BIN_ENV, [BIN_ENV, BIN_AGENT], environ)
    raise ValueError('Failed to exec ssh-agent')
  parentpostfork(rside, wside)
  my_agents=find_ssh_agents()

assert my_agents!=[], r'Agent could not be started.'

debug(ME_DIR, r'ME_DIR')
debug(SECRET, r'SECRET')
debug(my_agents, r'ssh agents')

# get Vault password
password=r''
if exists(SECRET) and isfile(SECRET):
  password=slurp(SECRET)
else:
  password=getpass(r'Vault password: ')
assert password!=r'', r'Empty password entered.'

# get fully decrypted JSON object from vault
vaultblob=slurp(r'vaults/keys.yml')
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
    environ.update({ 'SSH_AUTH_SOCK': agent })
    environ.update({ 'SSH_ASKPASS': WISP_PATH })
    environ.update({ 'DISPLAY': '' })
    if not fork():
      childpostfork(rside, wside);
      # Execute the desired program, replace the program image,
      # doesn't return
      execve(BIN_ADD, [BIN_ADD, key['path']], environ)
      raise ValueError('Failed to exec ssh-agent')
    parentpostfork(rside, wside)
...
