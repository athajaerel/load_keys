# load_keys

One file (plus one optionally) needed for this to work:

## vaults/keys.yml

    ---
    - keys:
      - name: key1
        path: ~/.ssh/id_rsa
        password: !vault | etc. encrypted password
      - name: another
        path: /add/as/many/as/you/like
        password: or_plaintext_for_the_foolhardy
    ...

## vaults/secret.txt (optional)

Maybe store the vault password in here, chmod'd 0400 for convenience... or not.
Add the `vault_password_file` option to ansible.cfg to automate it.

