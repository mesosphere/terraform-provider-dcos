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

{{< tf_arguments >}}
    {{< tf_arg name="path" required="true" desc="path to the secret." />}}
    {{< tf_arg name="value" required="true" desc="value of the secret." />}}
    {{< tf_arg name="store" default="default" desc="The name of the secret store." />}}
{{</ tf_arguments >}}
