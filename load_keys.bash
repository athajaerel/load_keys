#!/bin/bash
set -euo pipefail
SSH_SOCK=/tmp/ssh-agent-${USER}-screen
AGENT_LIST=$(/usr/bin/find /tmp/ssh-* -name agent\* -user ${USER} | /usr/bin/head -1)

[ -h ${SSH_SOCK} ] || /bin/ln -s ${AGENT_LIST} ${SSH_SOCK}
/usr/bin/ansible-playbook \
	$(/usr/bin/dirname $0)/load_keys.yml \
	-i $(/usr/bin/dirname $0)/inventory \
	-e ansible_python_interpreter=auto_legacy_silent \
	$*
