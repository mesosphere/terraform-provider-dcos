resource "dcos_job" "ajob" {
  name = "ajobid2"
  cmd  = "echo foobar"
  cpus = 1
  mem  = 64
  disk = 0
  docker_image = "ubuntu:18.04"
  description  = "the best description ever"
}
