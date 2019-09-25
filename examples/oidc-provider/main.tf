provider "dcos" {}

variable "client_id" {
  default = "Google Client ID"
}

variable "client_secret" {
  default = "Google Client Secret"
}

data "dcos_base_url" "current" {}

resource "dcos_security_cluster_oidc" "google" {
  provider_id = "google-idp"
  description = "Google"

  issuer   = "https://accounts.google.com"
  base_url = "https://${data.dcos_base_url.current.url}"

  client_id     = "${var.client_id}"
  client_secret = "${var.client_secret}"
}
