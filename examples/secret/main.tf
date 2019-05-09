provider "dcos" {}

resource "dcos_secret" "test1" {
  path = "foobar"
  value = "foobar1"
}

