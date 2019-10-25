---
title: "dcos_security_org_group_user"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_group_user
Provides a resource for assigning users into groups.

## Example Usage

```hcl
# Assign the bootstrap user into testgroup
provider "dcos" {
  cluster = "my-cluster"
}

resource "dcos_security_org_group" "testgroup" {
  gid         = "testgroup"
  description = "This group is for testing only"
}

resource "dcos_security_org_group_user" "testgroupassign" {
  gid = "${dcos_security_org_group.testgroup.gid}"
  uid = "bootstrapuser"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="gid" required="true" desc="Group ID." />}}
    {{< tf_arg name="uid" required="true" desc="User ID." />}}
{{</ tf_arguments >}}
