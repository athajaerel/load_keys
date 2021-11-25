#!/bin/bash
set -euo pipefail

ME_DIR=$(/usr/bin/dirname $0)
SSH_SOCK=/tmp/ssh-agent-${USER}-screen
AGENT_LIST=$(/usr/bin/find /tmp/ssh-* -name agent\* -user ${USER} | /usr/bin/head -1)
SECRET="${ME_DIR}/vaults/secret.txt"

if [ -e ${SECRET} ]; then
	SECRET="--vault-password-file ${SECRET}"
else
	SECRET=
fi

[ -h ${SSH_SOCK} ] || /bin/ln -s ${AGENT_LIST} ${SSH_SOCK}
/usr/bin/ansible-playbook \
	${ME_DIR}/load_keys.yml \
	-i ${ME_DIR}/inventory \
	-e ansible_python_interpreter=auto_legacy_silent \
	$*
