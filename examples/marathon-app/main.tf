provider "dcos" {}

resource "dcos_marathon_app" "app1" {
  name = "/app1"
  cmd  = "while true; do echo foo;sleep 30;done"

  # args = ["Hello World"]
  cpus = 0.1
  mem  = 64
}
