#!/bin/bash
/usr/bin/ansible-playbook ~/load_keys/load_keys.yml --vault-password-file=~/load_keys/vaults/secret.txt
