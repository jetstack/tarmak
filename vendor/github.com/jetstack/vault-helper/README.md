Vault Helper
============

This is tool is designed to automate PKI tasks for
[Tarmak](https://github.com/jetstack/tarmak) using Hashicorp's
[Vault](https://www.vaultproject.io>) as a backend.

`vault-helper` is designed to first run `setup`. This will ensure all CA
backends are mounted to the Vault server before applying roles and policies.
This process is idempotent.

`renew-token` ensures that if an init token is present in file, a new token will
be generated from that init token. This new token will be stored, deleting the
init token. If a token has already been generated, this token will be renewed
agaisnt the Vault server.

`cert` ensures that a private key has been generated and written to file. After
which `cert` will verify stored certificates against Vault. If unsucessful, will
issue a Certificate Signing Request to the Vault server using this private key.
The responding signed certificate is then stored at the given path.

`kubeconfig` will apply `cert` before encoding the certificates and private key
into a stored yaml file at the given path.

`setup`, `cert`, `read` and `kubeconfig` will all apply a `renew-token` with the
given token before continuing if successful.

`dev-server` is used only to set up a local development evnironment for testing.


vault-helper Usage
==================
```
Available Commands:
  cert        Create local key to generate a CSR. Call vault with CSR for specified cert role.
  dev-server  Run a vault server in development mode with kubernetes PKI created.
  help        Help about any command
  kubeconfig  Create local key to generate a CSR. Call vault with CSR for specified cert role. Write kubeconfig to yaml file.
  read        Read arbitrary vault path. If no output file specified, output to console.
  renew-token Renew token on vault server.
  setup       Setup kubernetes on a running vault server.
  version     Print the version number of vault-helper.

Flags:
  -h, --help            help for vault-helper
  -l, --log-level int   Set the log level of output. 0-Fatal 1-Info 2-Debug (default 1)

Use "vault-helper [command] --help" for more information about a command.
```

Vault Environment Variables
===========================
`vault-helper` requires the correct Vault environment variables to be set, for example:
```
$ export VAULT_ADDR=http://127.0.0.1:8200
```


Command Examples
=============================

### setup
```
$ vault-helper setup cluster-name
```


#### renew-token
```
$ vault-helper renew-token --init_role=cluster-name-master
```


### cert
```
$ vault-helper cert cluster-name/pki/k8s/sign/kube-apiserver k8s /etc/vault/name
```
