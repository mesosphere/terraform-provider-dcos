provider "dcos" {}

variable "app_id" {
  default = "jenkins"
}

resource "tls_private_key" "jenkins_service_account_private_key" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "dcos_security_org_service_account" "jenkins_service_account" {
  uid         = "${var.app_id}-principal"
  description = "Jenkins service account"
  public_key  = "${tls_private_key.jenkins_service_account_private_key.public_key_pem}"
}

locals {
  jenkins_principal_grants = [
    "dcos:mesos:master:framework:role:*",
    "dcos:mesos:master:task:user:nobody",
  ]
}

resource "dcos_security_org_user_grant" "testgrant" {
  count    = "${length(local.jenkins_principal_grants)}"
  uid      = "${dcos_security_org_service_account.jenkins_service_account.uid}"
  resource = "${element(local.jenkins_principal_grants, count.index)}"
  action   = "create"
}

# could be predefined data resource or special "service_account_secret"
locals {
  jenkins_secret = {
    scheme         = "RS256"
    uid            = "${dcos_security_org_service_account.jenkins_service_account.uid}"
    private_key    = "${tls_private_key.jenkins_service_account_private_key.private_key_pem}"
    login_endpoint = "https://master.mesos/acs/api/v1/auth/login"
  }
}

resource "dcos_security_secret" "jenkins-secret" {
  path = "${var.app_id}/jenkins-secret"

  value = "${jsonencode(local.jenkins_secret)}"
}

resource "dcos_package" "jenkins" {
  name   = "jenkins"
  app_id = "${var.app_id}"

  config_json = <<EOF
{"security":{"secret-name":"${dcos_security_secret.jenkins-secret.path}","strict-mode":true},"service":{"user":"nobody", "mem": 4096}}
EOF
}
