provider "dcos" {}

// Configure the cosmos repository to use
// (By default "universe" is used, here we are adding it for the sake of demo)
resource "dcos_package_repo" "universe" {
  name = "universe"
  url  = "https://universe.mesosphere.com/repo"
}

// Select the package version out of the cosmos repository you
// configured above
data "dcos_package_version" "kafka-zookeeper" {
  repo_url = "${dcos_package_repo.universe.url}"
  name     = "kafka-zookeeper"
  version  = "latest"
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
}

// You can chain more than one configurations
data "dcos_package_config" "kafka-zookeeper-kerberos" {
  extend = "${data.dcos_package_config.kafka-zookeeper.config}"

  section {
    path = "service.security.kerberos"

    map = {
      enabled               = true
      enabled_for_zookeeper = true
      keytab_secret         = "__dcos_base64__kafka_zookeeper_keytab"
      primary               = "kafka"
      realm                 = "LOCAL"
    }
  }

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
}
