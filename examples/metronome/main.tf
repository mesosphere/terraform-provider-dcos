resource "dcos_job" "ajob" {
  name = "ajobid"
  cmd  = "echo foo"
  cpus = 1
  mem  = 32
  disk = 0
  user = "root"
  description  = "the best description ever"
  max_launch_delay = 600

  docker {
    image = "ubuntu:latest"
  }

  placement_constraint {
    attribute = "host"
    operator = "LIKE"
  }

  env {
    this_is_not_a_key = "this_is_not_a_value"
    some_key = "some_val"
  }

  restart {
    active_deadline_seconds = 120
    policy = "NEVER"
  }

  artifacts {
    uri = "http://downloads.mesosphere.com/robots.txt"
    extract = false
    executable = true
    cache = false
  }

  volume {
    container_path = "/mnt/test"
    host_path = "/dev/null"
    mode = "RW"
  }
}

resource "dcos_job_schedule" "jobsched" {
  name = "${dcos_job.ajob.name}"
  cron = "5 * * * *"
}
