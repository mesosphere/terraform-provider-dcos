provider "dcos" {}

resource "dcos_security_secret" "env" {
  path  = "marathon-env-sercret"
  value = "pong"
}

resource "dcos_security_secret" "file" {
  path  = "marathon-file-sercret"
  value = "file-pong"
}

resource "dcos_marathon_app" "marathon-ping" {
  app_id    = "/marathon-test-ping"
  cpus      = 0.1
  mem       = 32
  instances = 1

  cmd = <<EOF
  cd $${SANDBOX_PATH};cat index2.html; echo $${MARATHON_ENV_SECRET} > index.html && python -m http.server $PORT0
EOF

  container {
    type = "MESOS"

    docker {
      image = "python:3"
    }

    volumes {
      container_path = "index2.html"
      secret         = "secret2"
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

  secrets {
    "MARATHON_ENV_SECRET" = "${dcos_security_secret.env.path}"
    "secret2"             = "${dcos_security_secret.file.path}"
  }

  port_definitions {
    protocol = "tcp"
    port     = 0
    name     = "pong-port"
  }

  require_ports = true
}
