provider "dcos" {}

// Configure the cosmos repository to use
// (By default "universe" is used, here we are adding it for the sake of demo)
resource "dcos_package_repo" "universe" {
  name = "Universe"
  url  = "https://universe.mesosphere.com/repo"
}

// Select the package version out of the cosmos repository you
// configured above
data "dcos_package_version" "kafka-zookeeper" {
  repo_url = "${dcos_package_repo.universe.url}"
  name     = "kafka-zookeeper"
  version  = "2.6.0-3.4.14"
}

// Provide some package configuration for the version you have selected
data "dcos_package_config" "kafka-zookeeper" {
  version_spec = "${data.dcos_package_version.kafka-zookeeper.spec}"

  section {
    path = "service"

    map {
      log_level = "INFO"
    }
  }

  section {
    path = "node"

    map {
      cpus      = 0.1
      mem       = 1024
      data_disk = 1024
      heap      = 512
    }
  }
}

// The package resource is responsible for installing the package you have
// specified and configured. Only installation-specific options are required
// in the resource itself.
resource "dcos_package" "kafka" {
  app_id = "zookeeper"
  config = "${data.dcos_package_config.kafka-zookeeper.config}"

  // By default the resource provider will wait until the service is found
  // available before continuing. You can configure this behavior with the
  // `wait` variable. (Default true)
  wait = true
}
