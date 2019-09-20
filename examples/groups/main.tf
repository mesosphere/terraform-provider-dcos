provider "dcos" {}

resource "dcos_security_org_group" "testgroup" {
  gid         = "testgroup"
  description = "This group is for testing only"
}

resource "dcos_security_org_group_user" "testgroupassign" {
  gid = "${dcos_security_org_group.testgroup.gid}"
  uid = "bootstrapuser"
}
