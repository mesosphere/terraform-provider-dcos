provider "dcos" {
  cluster = "soakhack"
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
