#!/bin/bash
[ -h /tmp/ssh-agent-${USER}-screen ] || /bin/ln -s $(/bin/find /tmp/ssh-* -name agent\* -user ${USER} | /bin/head -1) /tmp/ssh-agent-${USER}-screen
/usr/bin/ansible-playbook $(/bin/dirname $0)/load_keys.yml $*
