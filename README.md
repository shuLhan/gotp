# gotp

A command line interface to manage and generate Time-based One Time Password
(TOTP).

## SYNOPSIS

    gotp <command> <parameters...>


## COMMANDS

This section describe available command and its usage.

    add <LABEL> <HASH>:<BASE32-SECRET>[:DIGITS][:TIME-STEP][:ISSUER]

Add a TOTP secret identified by unique LABEL.
HASH is one of the valid hash function: SHA1, SHA256, or
SHA512.
BASE32-SECRET is the secret to generate one-time password
encoded in base32.
The DIGITS field is optional, define the number digits
generated for password, default to 6.
The TIME-STEP field is optional, its define the interval in
seconds, default to 30 seconds.
The ISSUER field is also optional, its define the name of
provider that generate the secret.

    export <FORMAT> [FILE]

Export all the issuers to file format that can be imported by provider.
Currently, the only supported FORMAT is "uri".
If FILE is not provided, it will print to the standard output.
The list of exported issuers are printed in order by its label.

    gen <LABEL> [N]

Generate N number passwords using the secret identified by LABEL.

    get <LABEL>

Get and print the issuer by its LABEL.
This will print the issuer secret, unencrypted.

    import <PROVIDER> <FILE>

Import the TOTP configuration from other provider.
Currently, the only supported PROVIDER is Aegis and the supported file
is .txt.

    list

List all labels stored in the configuration.

    remove <LABEL>

Remove LABEL from configuration.

    remove-private-key

Decrypt the issuer's value (hash:secret...) using current private key and
store it back to file as plain text.
The current private key will be removed from gotp directory.

    rename <LABEL> <NEW-LABEL>

Rename a LABEL into NEW-LABEL.

    set-private-key <PRIVATE-KEY-FILE>

Encrypt the issuer's value (hash:secret...) in the file using private key.
The supported private key is RSA.
Once completed, the PRIVATE-KEY-FILE will be copied to default user's gotp
directory, "$XDG_CONFIG_DIR/gotp/gotp.key".


##  ENCRYPTION

On the first run, the gotp command check for private key in the user's
configuration direction (see the private key location in FILES section).

The private key must be RSA based.

If the private key exist, all the OTP values (excluding the label) will be
stored as encrypted.

If the private key is not exist, the OTP configuration will be stored as
plain text.


##  FILES

$XDG_CONFIG_DIR/gotp:: Path to user's gotp directory.

$XDG_CONFIG_DIR/gotp/gotp.conf:: File where the configuration and
secret are stored.

$XDG_CONFIG_DIR/gotp/gotp.key:: Private key file to encrypt and decrypt the
issuer.

For Darwin/macOS the "$XDG_CONFIG_DIR" is equal to "$HOME/Library",
for Windows its equal to "%AppData%".


##  EXAMPLES

This section show examples on how to use gotp cli.

Add "my-totp" to configuration using SHA1 as hash function, "GEZDGNBVGY3TQOJQ"
as the secret, with 6 digits passwords, and 30 seconds as time step.

    $ gotp add my-totp SHA1:GEZDGNBVGY3TQOJQ:6:30


Generate 3 recent passwords from "my-totp",

    $ gotp gen my-totp 3
    gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
    847945
    326823
    767317


Import the exported Aegis TOTP from file,

    $ gotp import aegis aegis-export-uri.txt
    gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
    OK


List all labels stored in the configuration,

    $ gotp list
    gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
    my-totp


Remove a label "my-totp",

    $ gotp remove my-totp
    gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
    OK


Rename a label "my-totp" to "my-otp",

    $ gotp rename my-totp my-otp
    gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
    OK


##  Development

<https://git.sr.ht/~shulhan/gotp>:: Link to the source code.

<https://lists.sr.ht/~shulhan/gotp>:: Link to development and
discussion.

<https://todo.sr.ht/~shulhan/gotp>:: Link to submit an issue,
feedback, or request for new feature.

[Changelog](https://kilabit.info/project/gotp/CHANGELOG.html):: Change log
for each releases.
