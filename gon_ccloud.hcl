source = ["./dist/ccloud/darwin_amd64/ccloud"]
bundle_id = "io.confluent.cli.ccloud"

apple_id {
}

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/ccloud/darwin_amd64/ccloud_signed.zip"
}
