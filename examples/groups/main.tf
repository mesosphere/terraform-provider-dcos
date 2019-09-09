provider "dcos" {}

resource "dcos_iam_group" "testgroup" {
  gid         = "testgroup"
  description = "This group is for testing only"
}

resource "dcos_iam_group_user" "testgroupassign" {
  gid = "${dcos_iam_group.testgroup.gid}"
  uid = "bootstrapuser"
}
