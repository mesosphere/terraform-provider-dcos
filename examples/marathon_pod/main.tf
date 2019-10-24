provider "dcos" {}

resource "dcos_marathon_pod" "simplepod" {
  name = "simplepod"

  scaling {
    kind      = "fixed"
    instances = 1
  }

  container {
    name = "sleep1"

    exec {
      command_shell = "sleep 1000"
    }

    resources {
      cpus = 0.1
      mem  = 32
    }
  }

  network {
    mode = "HOST"
  }
}
