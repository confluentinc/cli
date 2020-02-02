source = ["./dist/confluent/darwin_amd64/confluent"]
bundle_id = "io.confluent.cli.confluent"

apple_id {
  username = "david.hyde@confluent.io"
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Confluent, Inc."
}

zip {
  output_path = "./dist/confluent/darwin_amd64/confluent_signed.zip"
}
