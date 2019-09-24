provider "dcos" {}

resource "tls_private_key" "test_service_account_key" {
  algorithm = "RSA"
  rsa_bits  = "2048"
}

resource "dcos_security_org_service_account" "test_service_account" {
  uid         = "testServiceAccount"
  description = "Terraform provider Test User"
  public_key  = "${tls_private_key.test_service_account_key.public_key_pem}"
}

resource "dcos_security_org_user_grant" "testgrant" {
  uid      = "${dcos_security_org_service_account.test_service_account.uid}"
  resource = "dcos:mesos:master:framework:role:kubernetes-role"
  action   = "create"
}
