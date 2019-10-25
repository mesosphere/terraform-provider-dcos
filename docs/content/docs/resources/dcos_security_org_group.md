---
title: "dcos_security_org_group"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_group
Provides a resource for creating DC/OS user groups.

## Example Usage

```hcl
# Create a group
provider "dcos" {
  cluster = "my-cluster"
}

resource "dcos_security_org_group" "testgroup" {
  gid         = "testgroup"
  description = "This group is for testing only"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="gid" required="true" desc="User ID to apply the grant on." />}}
    {{< tf_arg name="description" desc="a description for the group." />}}
{{</ tf_arguments >}}

## Attributes Reference
 addition to all arguments above, the following attributes are exported:

{{< tf_arguments >}}
    {{< tf_arg name="group_provider" desc="Group linked to an external provider." />}}
{{</ tf_arguments >}}
