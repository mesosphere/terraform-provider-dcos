---
title: "dcos_security_org_user"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_user
Provide a user resource. Managing users and their credentials.

## Example Usage

---

```hcl
# Create a Admin User with random password.
provider "dcos" {
  cluster = "my-cluster"
}

resource "random_password" "password" {
  length           = 16
  special          = true
  override_special = "_%@"
}

resource "dcos_security_org_user" "myadmin" {
  uid         = "myadmin"
  description = "Terraform managed admin user"
  password    = "${random_string.password.result}"
}

resource "dcos_security_group_user" {
  uid = "${dcos_security_org_user.myadmin.uid}"
  gid = "superusers"
}
```

## Argument Reference

---

The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="uid" required="true" desc="the user ID." />}}
    {{< tf_arg name="description" desc="a description for the User." />}}
    {{< tf_arg name="password" desc="Specified password for the User. Optional setting could also be maintained outside of Terraform." />}}
{{</ tf_arguments >}}
