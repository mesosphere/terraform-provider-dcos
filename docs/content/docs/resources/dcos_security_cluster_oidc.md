---
title: "dcos_security_cluster_oidc"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_cluster_oidc
Resource to maintain SAML authentication with DC/OS Cluster

## Example Usage

```hcl
# Assign the bootstrap user into testgroup
provider "dcos" {
  cluster = "my-cluster"
}

variable "client_id" {
  default = "Google Client ID"
}

variable "client_secret" {
  default = "Google Client Secret"
}

data "dcos_base_url" "current" {}

resource "dcos_security_cluster_oidc" "google" {
  provider_id = "google-idp"
  description = "Google"

  issuer   = "https://accounts.google.com"
  base_url = "https://${data.dcos_base_url.current.url}"

  client_id     = "${var.client_id}"
  client_secret = "${var.client_secret}"
}


```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="provider_id" required="true" desc="Unique Identifier for this Provider. Only lowercase characters allowed." />}}
    {{< tf_arg name="base_url" required="true" desc="The Clusters base URL." />}}
    {{< tf_arg name="description" desc="Description string for this provider." />}}
    {{< tf_arg name="client_id" required="true" desc="Client ID from identity provider." />}}
    {{< tf_arg name="client_secret" required="true" desc="Client secret from identity provider." />}}
    {{< tf_arg name="issuer" required="true" desc="Identity Provider issuer string." />}}
    {{< tf_arg name="ca_certs" desc="" />}}
    {{< tf_arg name="verify_server_certificate" default="false" desc="Verify SSL certificates." />}}
{{</ tf_arguments >}}
