source = ["./dist/macos_darwin_amd64/terraform-provider-dcos"]
bundle_id = "com.mesosphere.terraform-provider-dcos"

# AC_USERNAME and AC_PASSWORD must be set

sign {
  application_identity = "Developer ID Application: Mesosphere Inc. (JQJDUUPXFN)"
}
