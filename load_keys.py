#!/usr/bin/env python3

from sys import path
from os import environ, stat, symlink, execve
from os.path import dirname, realpath, exists, isdir
from stat import S_ISSOCK
from glob import glob
from pwd import getpwuid

from config import *

USER=environ.get('USER')
ME_DIR=path[0]
SECRET=realpath('%s/vaults/secret.txt' % ME_DIR)

def debug(line, prefix=None):
  if (DEBUG_MODE):
    if (prefix):
      print('%s: %s' % (prefix, line))
    else:
      print(line)

def file_owner(filename):
  return getpwuid(stat(filename).st_uid).pw_name

def issock(filename):
  return S_ISSOCK(stat(filename).st_mode)

def find_ssh_agents():
  # find all ssh agents
  agent_dirs=glob('%s/ssh-*' % TMPDIR)
  my_agents=[]
  for agent_dir in agent_dirs:
    # must be a directory not a symlink
    if not isdir(agent_dir):
      continue
    # must be owned by user
    if not file_owner(agent_dir) == USER:
      continue
    # must have agent.* file
    agent_files=glob('%s/agent.*' % (agent_dir))
    for agent_file in agent_files:
      # must be owned, and a socket file
      if not issock(agent_file):
        continue
      if not file_owner(agent_file) == USER:
        continue
      my_agents+=[agent_file]
  return my_agents

extra_opts=''
if exists(SECRET):
  extra_opts+='--vault-password-file %s' % SECRET

my_agents=find_ssh_agents()
if my_agents==[]:
  execve('/usr/bin/env', ['ssh-agent'])
  my_agents=find_ssh_agents()

debug(extra_opts, 'extra_opts')
debug(ME_DIR, 'ME_DIR')
debug(SECRET, 'SECRET')
debug(my_agents, 'ssh agents')

#- name: 'Add keys to all owned ssh caches'
#  vars:
#    wisp_full_path: '/dev/shm/wisp.bash'
#  vars_files:
#    - 'vaults/keys.yml'
#  gather_facts: no
#  tasks:
#
#    # Get the keys into all correctly owned ssh agents
#    # Default to user 'nobody' as this will find no agents and default
#    # to a no-op, rather than silently run with '' and do who knows
#    # what
#    - name: 'Find all agents owned by this user'
#      register: sshAuthSock
#      become: no
#      failed_when: False
#      command:
#        argv:
#          - 'find'
#          - '/tmp'
#          - '-name'
#          - 'agent.*'
#          - '-user'
#          - '{{ lookup("env", "USER") | default("nobody") }}'
#
#    # TO DO: if no results, run ssh-agent and use that value in nest 0
#
#    - name: 'Load keys into agents'
#      with_nested:
#        - '{{ sshAuthSock.stdout_lines }}'
#        - '{{ keys }}'
#      include_tasks: 'block_load_keys.yml'

