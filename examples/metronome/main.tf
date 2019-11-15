data "dcos_job" "job" {
  name = "somejob"
}

resource "dcos_job" "ajob" {
  name             = "ajobid"
  cmd              = "echo foo"
  cpus             = 1
  mem              = 64
  disk             = 0
  user             = "root"
  description      = "the best description ever"
  max_launch_delay = 600

  docker {
    image = "ubuntu:latest"
  }

  placement_constraint {
    attribute = "host"
    operator  = "LIKE"
  }

  env {
    key    = "my_env_key"
    secret = "secret1"
  }

  env {
    key   = "cool_key"
    value = "cool_value"
  }

  secrets {
    secret1     = "/something"
    cool_secret = "something_else"
  }

  restart {
    active_deadline_seconds = 120
    policy                  = "NEVER"
  }

  artifacts {
    uri        = "http://downloads.mesosphere.com/robots.txt"
    extract    = false
    executable = true
    cache      = false
  }

  volume {
    container_path = "/mnt/test"
    host_path      = "/dev/null"
    mode           = "RW"
  }
}

resource "dcos_job_schedule" "jobsched" {
  dcos_job_id = "${dcos_job.ajob.name}"
  name        = "someschedule"
  cron        = "0,30 * * * *"
}

output "somejob_name" {
  value = "${data.dcos_job.job.name}"
}

output "somejob_cpu" {
  value = "${data.dcos_job.job.cpus}"
}

output "somejob_mem" {
  value = "${data.dcos_job.job.mem}"
}

output "somejob_disk" {
  value = "${data.dcos_job.job.disk}"
}

output "somejob_cmd" {
  value = "${data.dcos_job.job.cmd}"
}
