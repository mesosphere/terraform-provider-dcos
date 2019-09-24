provider "dcos" {}

resource "dcos_security_secret" "test1" {
  path  = "foobar"
  value = "foobar1"
}
