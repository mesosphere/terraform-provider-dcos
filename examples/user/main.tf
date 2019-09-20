provider "dcos" {}

resource "dcos_security_org_user" "testuser" {
  uid         = "testuserpw"
  description = "a test user with password set"
  password    = "mysecurepassword"
}
