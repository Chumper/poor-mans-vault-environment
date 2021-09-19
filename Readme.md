# Poor mans vault environment aka "PMVE"

At my company we are using a multitenant kubernetes cluster and are storing our secrets in Vault.

When I was looking at solutions on how to inject Vault secrets into pods I found the solution done in the Banzai Cloud: https://banzaicloud.com/blog/inject-secrets-into-pods-vault-revisited/

Unfortunately this requires a mutating admission webhook which is not possible when you only have a namespace on a kubernetes cluster and no cluster admin rights.

I therefore decided to implement a poor mans solution for this problem.

# Usage

`pmve` is intended to be used as process executor for the original process.  
`pmve` will simply look for any environment variable starting with `vault:` and replaces the variable with the content of the secret.

E.g. `TEST=vault:/secret/my-secret#password` will be replaced to `TEST=my-secret-password` assuming that in vault there is a secret at the given path.

`pmve` will then replace itself with the process given in the args.

## Quick start

```bash
Usage of pvme:
  -keep-secrets
        True if Vault secrets should not be deleted in the environment
  -vault-addr string
        The Vault address (default "http://localhost:8200")
  -vault-role-id string
        The role ID of the Vault approle
  -vault-secret-id string
        The secret ID of the Vault approle
```

Example:
