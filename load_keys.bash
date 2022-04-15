#!/bin/bash
set -euo pipefail

ME_DIR=$(/usr/bin/dirname $0)
SECRET="${ME_DIR}/vaults/secret.txt"

if [ -e ${SECRET} ]; then
	SECRET="--vault-password-file ${SECRET}"
else
	SECRET=
fi

/usr/bin/ansible-playbook ${ME_DIR}/load_keys.yml ${SECRET}
