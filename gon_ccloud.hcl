source = ["./dist/ccloud/ccloud_darwin_amd64/ccloud"]
bundle_id = "io.confluent.cli.ccloud"

apple_id {
}

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/ccloud/ccloud_darwin_amd64/ccloud_signed.zip"
}
