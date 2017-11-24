# vault-helper usage
```
Automates PKI tasks using Hashicorp's Vault as a backend.

Usage:
  vault-helper [command]

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

## Vault helper requires the following environment variable set
Export vault address:
```
$ export VAULT_ADDR=http://127.0.0.1:8200
```


## Vault Helper command examples
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
