# load_keys

Two (possibly three) files needed for this to work:

## vars/keys.yml

    ---
    - keys:
      - name: key1
        path: ~/.ssh/id_rsa
      - name: another
        path: /add/as/many/as/you/like
    ...

## vaults/key_passwords.yml

    ---
    - passwords:
        key1: !vault | etc. encrypted password
        another: or_plaintext_for_the_foolhardy
    ...

## vaults/secret.txt (optional)

Maybe store the vault password in here, chmod'd 0400 for convenience... or not.
Add the `vault_password_file` option to ansible.cfg to automate it.

