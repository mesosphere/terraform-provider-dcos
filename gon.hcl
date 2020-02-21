source = ["./dist/terraform-provider-dcos_darwin_amd64/terraform-provider-dcos"]
bundle_id = "com.mesosphere.terraform-provider-dcos"

# AC_USERNAME and AC_PASSWORD must be set

sign {
  application_identity = "Developer ID Application: Mesosphere Inc. (JQJDUUPXFN)"
}

zip {
  output_path = "dist/terraform-provider-dcos-darwin-amd64.zip"
}
