---
title: "dcos_security_org_external_user"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_external_user
Resource to maintain SAML authentication with DC/OS Cluster

## Example Usage

```hcl
provider "dcos" {
  cluster = "my-cluster"
}

locals {
  user_list = ["jdoe@mesosphere.com"]
}

data "dcos_base_url" "current" {}

resource "dcos_security_cluster_oidc" "google" {
  provider_id = "google-idp"
  description = "Google"

  issuer   = "https://accounts.google.com"
  base_url = "${data.dcos_base_url.current.url}"

  client_id     = "<...>"
  client_secret = "<...>"
}

resource "dcos_security_org_external_user" "soakusers" {
  count         = "${length(local.user_list)}"
  uid           = "${element(local.user_list,count.index)}"
  description   = "Terraform managed OIDC Users"
  provider_id   = "${dcos_security_cluster_oidc.google.provider_id}"
  provider_type = "oidc"
}

resource "dcos_security_org_group_user" "soakusergroups" {
  count = "${length(local.user_list)}"
  uid   = "${element(local.user_list,count.index)}"
  gid   = "superusers"

  depends_on = ["dcos_security_org_external_user.soakusers"]
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="uid" required="true" desc="ID of the account is used by default." />}}
    {{< tf_arg name="provider_id" required="true" desc="Provider ID for this external user e.g. OneLogin" />}}
    {{< tf_arg name="provider_type" required="true" desc="Type of external provider. (`ldap` or `oidc` or `saml`)" />}}
    {{< tf_arg name="description" desc="Description of the newly created external user." />}}
{{</ tf_arguments >}}
