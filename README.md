# gotp

A command line interface to manage and generate Time-based One Time Password
(TOTP).

## SYNOPSIS

```
gotp <command> <parameters...>
```

## DESCRIPTION

```
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

    Decrypt the issuer's value (hash:secret...) using previous private key and
    store it back to file as plain text.

rename <LABEL> <NEW-LABEL>

	Rename a LABEL into NEW-LABEL.

set-private-key <PRIVATE-KEY-FILE>

    Encrypt the issuer's value (hash:secret...) in the file using private key.
    The supported private key is RSA.
```

##  ENCRYPTION

On the first run, the gotp command will ask for path of private key.
If the key exist, all the OTP values (excluding the label) will be encrypted.
The private key must be RSA based.

One can skip inputting the private key by pressing enter, and the OTP
configuration will be stored as plain text.

##  FILES

$USER_CONFIG_DIR/gotp/gotp.conf:: Path to file where the configuration and
secret are stored.

##  EXAMPLES

Add "my-totp" to configuration using SHA1 as hash function, "GEZDGNBVGY3TQOJQ"
as the secret, with 6 digits passwords, and 30 seconds as time step.

```
$ gotp add my-totp SHA1:GEZDGNBVGY3TQOJQ:6:30
```

Generate 3 recents passwords from "my-totp",

```
$ gotp gen my-totp 3
gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
847945
326823
767317
```

Import the exported Aegis TOTP from file,

```
$ gotp import aegis aegis-export-uri.txt
gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
OK
```

List all labels stored in the configuration,

```
$ gotp list
gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
my-totp
```

Remove a label "my-totp",

```
$ gotp remove my-totp
gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
OK
```

Rename a label "my-totp" to "my-otp",

```
$ gotp rename my-totp my-otp
gotp: reading configuration from /home/$USER/.config/gotp/gotp.conf
OK
```
