provider "dcos" {}

resource "dcos_marathon_app" "marathon-ping" {
  app_id    = "/marathon-ping"
  cpus      = 0.1
  mem       = 32
  instances = 1

  cmd = <<EOF
echo "pong" > index.html && python -m http.server $PORT0
EOF

  container {
    type = "DOCKER"

    docker {
      image = "python:3"
    }
  }

  health_checks {
    path                     = "/"
    protocol                 = "MESOS_HTTP"
    port_index               = 0
    grace_period_seconds     = 5
    interval_seconds         = 10
    timeout_seconds          = 10
    max_consecutive_failures = 3
  }

  port_definitions {
    protocol = "tcp"
    port     = 0
    name     = "pong-port"
  }

  require_ports = true
}
