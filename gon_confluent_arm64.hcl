source = ["./dist/confluent/signed_darwin_arm64/confluent"]
bundle_id = "io.confluent.cli.confluent"

apple_id {
}

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/confluent/confluent_darwin_arm64/confluent_signed.zip"
}
