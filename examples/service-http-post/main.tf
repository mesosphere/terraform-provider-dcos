provider "dcos" {}

locals {
  github_user  = "some-user"
  github_token = "some-token"
}

resource "dcos_service_http_request" "jenkins_credentials" {
  service_name = module.jenkins.service_name
  path         = "/credentials/store/system/domain/_/createCredentials"
  method       = "POST"

  header {
    name  = "Content-Type"
    value = "application/x-www-form-urlencoded"
  }

  body = "json={\"credentials\":{\"scope\":\"GLOBAL\", \"username\":\"'${local.github_user}'\", \"password\":\"'${local.github_token}'\", \"id\":\"GitHubUserWithToken\", \"description\":\"GitHub user and pass/token to download repositories \", \"stapler-class\":\"com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl\"}}&Submit=OK"
}
