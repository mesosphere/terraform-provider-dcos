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
  jenkins_principal_grants_create = [
    "dcos:mesos:master:framework:role:*",
    "dcos:mesos:master:task:user:nobody",
  ]
}

resource "dcos_security_org_user_grant" "jenkin_grant_create" {
  count    = "${length(local.jenkins_principal_grants_create)}"
  uid      = "${dcos_security_org_service_account.jenkins_service_account.uid}"
  resource = "${element(local.jenkins_principal_grants_create, count.index)}"
  action   = "create"
}

locals {
  jenkins_principal_grants_read = [
    "dcos:secrets:list:default:/${var.app_id}",
  ]
}

resource "dcos_security_org_user_grant" "jenkin_grant_read" {
  count    = "${length(local.jenkins_principal_grants_read)}"
  uid      = "${dcos_security_org_service_account.jenkins_service_account.uid}"
  resource = "${element(local.jenkins_principal_grants_read, count.index)}"
  action   = "read"
}

locals {
  jenkins_principal_grants_full = [
    "dcos:secrets:default:/${var.app_id}/*",
  ]
}

resource "dcos_security_org_user_grant" "jenkin_grant_full" {
  count    = "${length(local.jenkins_principal_grants_full)}"
  uid      = "${dcos_security_org_service_account.jenkins_service_account.uid}"
  resource = "${element(local.jenkins_principal_grants_full, count.index)}"
  action   = "full"
}

locals {
  jenkins_secret = {
    scheme         = "RS256"
    uid            = "${dcos_security_org_service_account.jenkins_service_account.uid}"
    private_key    = "${tls_private_key.jenkins_service_account_private_key.private_key_pem}"
    login_endpoint = "https://master.mesos/acs/api/v1/auth/login"
  }
}

data "dcos_package_version" "jenkins" {
  name     = "jenkins"
  version  = "latest"
}
resource "dcos_security_secret" "jenkins-secret" {
  path = "${var.app_id}/service-account-secret"

  value = "${jsonencode(local.jenkins_secret)}"
}

data "dcos_package_config" "jenkins" {
  version_spec = "${data.dcos_package_version.jenkins.spec}"
  autotype          = true

  section {
    path = "service"
    map = {
        user = "nobody"
    }
  }

  section {
    path = "security"
    map = {
        secret-name = "${dcos_security_secret.jenkins-secret.path}"
        strict-mode = "true"
    }
  }
}

resource "dcos_package" "jenkins" {
  app_id = "${var.app_id}"
  config = "${data.dcos_package_config.jenkins.config}"
}
