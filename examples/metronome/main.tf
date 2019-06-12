resource "dcos_job" "ajob" {
  name = "ajobid"
  cmd  = "echo foo"
  cpus = 1
  mem  = 32
  disk = 0
  description  = "the best description ever"

  docker {
    image = "ubuntu:latest"
  }

  artifacts {
    uri = "https://s3.amazonaws.com/soak-clusters/artifacts/soak113s/logs-elasticsearch-indices-rotate2.sh"
    extract = false
    executable = true
    cache = false
  }
}
