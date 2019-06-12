resource "dcos_job" "ajob" {
  name = "ajobid2"
  cmd  = "echo foo"
  cpus = 1
  mem  = 32
  disk = 0
  docker_image = "ubuntu:latest"
  description  = "the best description ever"
}
