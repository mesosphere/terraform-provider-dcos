provider "dcos" {}

data "dcos_token" "current" {}
data "dcos_base_url" "current" {}

output "dcos_token" {
  description = "current DC/OS token"
  value       = "${data.dcos_token.current.token}"
}

output "dcos_base_url" {
  description = "current DC/OS Base URL"
  value       = "${data.dcos_base_url.current.url}"
}
