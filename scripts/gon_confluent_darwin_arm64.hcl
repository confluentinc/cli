source = ["./dist/confluent-darwin-arm64_darwin_arm64/confluent"]
bundle_id = "io.confluent.cli.confluent"

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/confluent-darwin-arm64_darwin_arm64/confluent_signed.zip"
}