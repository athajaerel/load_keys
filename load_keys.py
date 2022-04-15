#!/usr/bin/env python3

from sys import path
from os import environ, stat, symlink, execve
from os.path import dirname, realpath, exists, isdir, islink
from glob import glob
from pwd import getpwuid

DEBUG_MODE=True
USER=environ.get('USER')
ME_DIR=path[0]
TMPDIR='/tmp'
SECRET=realpath('%s/vaults/secret.txt' % ME_DIR)

def debug(line):
  if (DEBUG_MODE):
    print(line)

def file_owner(filename):
  return getpwuid(stat(filename).st_uid).pw_name

def find_owned_agent():
  GLOB_DIR='%s/ssh-*' % TMPDIR
  # find all ssh agents
  agents=glob(GLOB_DIR)
  # find owned ssh agent
  my_agent=''
  for agent in agents:
    debug(agent)
    # must be a directory not a symlink
    if not isdir(agent):
      continue
    # must be owned
    debug('dir')
    if file_owner(agent) == USER:
      debug('owned')
      # must contain agent.* file
      glob_file='%s/agent.*' % (agent)
      debug(glob_file)
      agent_files = glob(glob_file)
      for agent_file in agent_files:
        return agent_file
  raise ValueError('No ssh-agent found')

debug(ME_DIR)
debug(SECRET)

extra_opts=''
if exists(SECRET):
  extra_opts+='--vault-password-file %s' % SECRET

debug(extra_opts)

try:
  my_agent=find_owned_agent()
except:
  execve('/usr/bin/env', ['ssh-agent'])
  my_agent=find_owned_agent()

debug(my_agent)


#/usr/bin/ansible-playbook ${ME_DIR}/load_keys.yml ${SECRET}

#---
#- name: 'Add keys to all owned ssh caches'
#  hosts: 'localhost'
#  connection: 'local'
#  run_once: yes
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
#...

