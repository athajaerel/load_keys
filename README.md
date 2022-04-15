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
If the secret file exists it will be used automatically.

## Now with much faster Python version

The Ansible version takes several seconds, which is inconvenient if it's in `.bashrc`. If you open terminals in quick succession, you get race condition issues.

The Python version uses the same files but executes in around a quarter of a second. That's much better. Though I might see if I can shave that down a little more.

Asks for the vault password if not supplied in `secret.txt`.

    $ ./load_keys.py
