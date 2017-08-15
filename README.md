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
  version     Print the version number of vault-helper

Flags:
  -p, --config-path string   Set config path to directory with tokens (default "/etc/vault")
  -h, --help                 help for vault-helper

Use "vault-helper [command] --help" for more information about a command.
```

## Vault helper requires the following environment variables set
Export root token to environment variable:
```
$ export VAULT_TOKEN=########-####-####-####-############
```
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
$ vault-helper renew-token cluster-name --role=admin
```

### cert
```
$ vault-helper cert cluster-name/pki/k8s/sign/kube-apiserver k8s /etc/vault/name
```
