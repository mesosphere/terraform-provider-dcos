---
title: "dcos_security_org_service_account"
variant: enterprise
type: docs
weight: 4
---

# Resource: dcos_security_org_service_account
Provides a resource for creating service accounts.

## Example Usage

```hcl
# Create a Service Account from a generated private key
provider "dcos" {
  cluster = "my-cluster"
}

resource "tls_private_key" "k8s" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "dcos_security_org_service_account" "k8s-sa" {
  uid         = "kubernetes-service-account"
  description = "Terraform provider Test User"
  public_key  = "${tls_private_key.k8s.public_key_pem}"
}

resource "dcos_security_org_user_grant" "k8s-grant" {
  uid      = "${dcos_security_org_service_account.k8s-sa.uid}"
  resource = "dcos:mesos:master:framework:role:kubernetes-role"
  action   = "create"
}

```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="uid" required="true" desc="User ID to apply the grant on." />}}
    {{< tf_arg name="description" desc="a description for the Service Account." />}}
    {{< tf_arg name="public_key" required="true" desc="Public key to use." />}}
{{</ tf_arguments >}}
