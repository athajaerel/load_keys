#!/usr/bin/env python3

from sys import path
from os import environ, stat, symlink, execve
from os.path import dirname, realpath, exists, isdir, isfile
from stat import S_ISSOCK
from glob import glob
from pwd import getpwuid
from ansible.constants import DEFAULT_VAULT_ID_MATCH # requires ?
from ansible.parsing.vault import VaultLib, VaultSecret # requires ?
from getpass import getpass
from json import loads, dumps
from yaml import load, BaseLoader, YAMLError # requires pyyaml

from config import *

USER=environ.get(r'USER')
ME_DIR=path[0]
SECRET=realpath(r'%s/vaults/secret.txt' % ME_DIR)
WISP_PATH=r'/dev/shm/wisp.bash'

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

def isjson(myblob):
  try:
    loads(myblob)
  except ValueError as e:
    return False
  return True

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
  assert isyaml(myblob)==True
  myyaml=loadyaml(myblob)
  return myyaml

def decryptfield(node, password):
  # decrypt all strings beginning $ANSIBLE_VAULT;1.1;AES256
  splitnode=node.split('\n', 1)
  if (splitnode[0]=='$ANSIBLE_VAULT;1.1;AES256'):
    v = VaultLib([(DEFAULT_VAULT_ID_MATCH, VaultSecret(password.encode()))])
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

my_agents=find_ssh_agents()
if my_agents==[]:
  execve(r'/usr/bin/env', [r'ssh-agent'])
  my_agents=find_ssh_agents()

assert my_agents!=[]

debug(ME_DIR, r'ME_DIR')
debug(SECRET, r'SECRET')
debug(my_agents, r'ssh agents')

# get Vault password
password=r''
if exists(SECRET) and isfile(SECRET):
  password=slurp(SECRET)
else:
  password=getpass(r'Vault password: ')
assert password!=r''

# get fully decrypted JSON object from vault
vaultblob=slurp(r'vaults/keys.yml')
assert vaultblob!=r''
debug(vaultblob, 'vault blob')

myobj=None
# if it's a JSON file, load it as JSON and decrypt the password entries
if isyaml(vaultblob):
  myobj=objectfromyaml(vaultblob)
  myobj=jsonwalk(myobj, decryptfield, password)
else:
  # otherwise, decrypt as a blob and confirm it's JSON
  # TODO
  vault=Vault(password)
  myobj=vault.load(vaultblob)

assert myobj!=None
debug(myobj)

#- name: 'Add keys to all owned ssh caches'
#  vars:
#    wisp_full_path: '/dev/shm/wisp.bash'
#  vars_files:
#    - 'vaults/keys.yml'
#
#    # TO DO: if no results, run ssh-agent and use that value in nest 0
#    - name: 'Load keys into agents'
#      with_nested:
#        - '{{ sshAuthSock.files | selectattr("pw_name", "eq", lookup("env", "USER") | default("nobody")) }}'
#        - '{{ keys }}'
#      include_tasks: 'block_load_keys.yml'
