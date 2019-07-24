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

  // Each section installs configuration values to a designated
  // location in the configuration.
  section {
    path = "service"

    map {
      log_level = "INFO"
    }
  }

  // Multiple sections can appear. You can even use raw JSON if required
  section {
    path = "node"

    json = <<EOF
{
  "cpus": 0.5,
  "mem": 1024,
  "data_disk": 1024,
  "heap": 512
}
EOF
  }
}

// You can chain configuration blocks
data "dcos_package_config" "kafka-zookeeper-kerberos" {
  // Point to the previous configuration's `config` field
  extend = "${data.dcos_package_config.kafka-zookeeper.config}"

  // New sections will be appended to the previous configuration.
  // In case of a merge collision, the newer values will be considered
  section {
    path = "service.security.kerberos.kdc"

    map {
      hostname = "kdc.marathon.autoip.dcos.thisdcos.directory"
      port     = 2500
    }
  }
}

// The package resource is responsible for installing the package you have
// specified and configured. Only installation-specific options are required
// in the resource itself.
resource "dcos_package" "kafka" {
  app_id = "zookeeper"
  config = "${data.dcos_package_config.kafka-zookeeper-kerberos.config}"

  // By default the resource provider will wait until the service is found
  // available before continuing. You can configure this behavior with the
  // `wait` variable. (Default true)
  wait = true
}
