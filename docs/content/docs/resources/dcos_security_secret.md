---
title: "dcos_security_secret"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_secret
Provides an secret resource. This allows to maintain secrets in a DC/OS secret store.

## Example Usage

---

```hcl
# Create a Secret containing a random password
provider "dcos" {
  cluster = "my-cluster"
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "_%@"
}

resource "dcos_security_secret" "myapp-password" {
  path  = "/myapp/password"
  value = "${random_string.password.result}"
}
```

## Argument Reference

---

The following arguments are supported

- `path` (Required) path the to secret.
- `value` (Required) value of the secret.
- `store` (Optional) The name of the secret store. Defaults to `default`
