provider "dcos" {}

variable "app_id" {
  default = "kubernetes"
}

resource "tls_private_key" "service_account_key" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "dcos_security_org_service_account" "service_account" {
  uid         = "${var.app_id}-principal"
  description = "Kubernets Service Account"
  public_key  = "${tls_private_key.service_account_key.public_key_pem}"
}

locals {
  principal_create_grants = [
    "dcos:mesos:master:reservation:role:kubernetes-role",
    "dcos:mesos:master:framework:role:kubernetes-role",
    "dcos:mesos:master:task:user:nobody",
  ]
}

resource "dcos_security_org_user_grant" "principal_create_grants" {
  count    = "${length(local.principal_create_grants)}"
  uid      = "${dcos_security_org_service_account.service_account.uid}"
  resource = "${element(local.principal_create_grants, count.index)}"
  action   = "create"
}

# could be predefined data resource or special "service_account_secret"
locals {
  jenkins_secret = {
    scheme         = "RS256"
    uid            = "${dcos_security_org_service_account.service_account.uid}"
    private_key    = "${tls_private_key.service_account_key.private_key_pem}"
    login_endpoint = "https://master.mesos/acs/api/v1/auth/login"
  }
}

resource "dcos_secret" "secret" {
  path = "${var.app_id}/sa"

  value = "${jsonencode(local.jenkins_secret)}"
}

resource "dcos_package" "package" {
  name   = "kubernetes"
  app_id = "${var.app_id}"

  config_json = <<EOF
{"service":{"service_account": "${dcos_iam_service_account.service_account.uid}", "service_account_secret": "${dcos_secret.secret.path}"}}
EOF

  depends_on = ["dcos_security_org_user_grant.principal_create_grants"]
}
