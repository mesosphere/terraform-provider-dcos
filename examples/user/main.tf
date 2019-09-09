provider "dcos" {}

resource "dcos_iam_user" "testuser" {
  uid         = "testuserpw"
  description = "a test user with password set"
  password    = "mysecurepassword"
}
