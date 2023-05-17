source = ["./dist/confluent-darwin-amd64-homebrew_darwin_amd64_v1/confluent"]
bundle_id = "io.confluent.cli.confluent"

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/confluent-darwin-amd64-homebrew_darwin_amd64_v1/confluent_signed.zip"
}