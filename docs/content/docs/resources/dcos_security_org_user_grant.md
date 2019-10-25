---
title: "dcos_security_org_user_grant"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_user_grant
Provides a grant resource maintaining a grant on a User or Service Account.

## Example Usage

```hcl
# Create a Secret containing a random password
provider "dcos" {
  cluster = "my-cluster"
}

resource "dcos_security_org_user" "myadmin" {
  uid         = "myadmin"
  description = "Terraform managed admin user"
}

locals {
  admin_full_grants = [
    "dcos:adminrouter:service:marathon",
    "dcos:adminrouter:ops:slave"
  ]
}

resource "dcos_security_org_user_grant" "myadmin-full-grants" {
  count    = "${length(local.admin_full_grants)}"
  uid      = "${dcos_security_org_user.myadmin.uid}"
  resource = "${element(local.admin_full_grants, count.index)}"
  action   = "full"
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="uid" required="true" desc="User ID to apply the grant on." />}}
    {{< tf_arg name="resource" required="true" desc="resource to grant access." />}}
    {{< tf_arg name="action" required="true" desc="granted action on resource." />}}
{{</ tf_arguments >}}
