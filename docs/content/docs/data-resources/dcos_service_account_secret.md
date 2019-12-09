---
title: "dcos_service_account_secret"
type: docs
weight: 5
---

# Data Resource: dcos_service_account_secret
Computes the contents for the service account secret

## Example Usage
```hcl
# Create a private key using an external provider (eg. tls)
resource "tls_private_key" "service_account_key" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

# Pass the private key and the user account to the resource
data "dcos_service_account_secret" "service_account" {
    uid            = "user-login"
    private_key    = "${tls_private_key.service_account_key.private_key_pem}"
}

# Handle the contents (eg. upload to a secret)
resource "dcos_security_secret" "service_account_secret" {
  path = "my-service/service-account"
  value = "${data.dcos_service_account_secret.contents}"
}
```

## Argument Reference
The following arguments are supported

{{< tf_arguments >}}
    {{< tf_arg name="private_key" required="true" >}}
        The PEM-encoded contents of a private key. This can be either a PKCS1 private key or PKCS8 private key without password. Any other type will be rejected.
    {{</ tf_arg >}}
    {{< tf_arg name="uid" required="true" desc="The user ID." />}}
    {{< tf_arg name="login_endpoint" required="false" default="https://leader.mesos/acs/api/v1/auth/login" desc="Override the default login endpoint that will be used by the service." />}}
    {{< tf_arg name="contents" output="true" >}}
        This is an output (read-only) variable with the contents of the service account secret. This value can be safely uploaded to a service account secret and later used by the service in DC/OS.
    {{</ tf_arg >}}
{{</ tf_arguments >}}
