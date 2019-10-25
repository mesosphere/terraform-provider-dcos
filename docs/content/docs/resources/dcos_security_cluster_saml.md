---
title: "dcos_security_cluster_saml"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_cluster_saml
Resource to maintain SAML authentication with DC/OS Cluster

## Example Usage

```hcl
# Assign the bootstrap user into testgroup
provider "dcos" {
  cluster = "my-cluster"
}

data "dcos_base_url" "current" {}

resource "dcos_security_cluster_saml" "OneloginTest" {
  provider_id = "onelogin"
  description = "OneLogin SAML Provider"

  # SAML provider metadata from a file
  idp_metadata = "${file("~/testcluster-onelogin.xml")}"
  base_url     = "${data.dcos_base_url.current.url}"
}

output "sp_metadata" {
  value = "${dcos_iam_saml_provider.OneloginTest.metadata}"
}

output "callback_url" {
  value = "${dcos_iam_saml_provider.OneloginTest.callback_url}"
}

output "entity_id" {
  value = "${dcos_iam_saml_provider.OneloginTest.entity_id}"
}


```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="provider_id" required="true" desc="Unique Identifier for this Provider. Only lowercase characters allowed." />}}
    {{< tf_arg name="idp_metadata" required="true" desc="IDP Metadata." />}}
    {{< tf_arg name="description" desc="Description string for this provider." />}}
    {{< tf_arg name="base_url" desc="Service provider base URL." />}}
{{</ tf_arguments >}}

## Attributes Reference
 addition to all arguments above, the following attributes are exported:

{{< tf_arguments >}}
    {{< tf_arg name="callback_url" desc="SAML Callbackurl." />}}
    {{< tf_arg name="metadata" desc="SAML service provider metadata." />}}
    {{< tf_arg name="entity_id" desc="Provided entity ID." />}}
{{</ tf_arguments >}}
