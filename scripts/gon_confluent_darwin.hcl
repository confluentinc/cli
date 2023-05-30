source = ["path/to/confluent"]
bundle_id = "io.confluent.cli.confluent"

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "path/to/confluent_signed.zip"
}