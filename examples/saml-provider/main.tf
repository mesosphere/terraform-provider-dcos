provider "dcos" {}

resource "dcos_security_cluster_saml" "OneloginTest" {
  provider_id  = "onelogin"
  description  = "OneLogin SAML Provider changed"
  idp_metadata = "${file("~/testcluster-onelogin.xml")}"
  base_url     = "https://julfertsv02-815177521.us-east-1.elb.amazonaws.com"
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
