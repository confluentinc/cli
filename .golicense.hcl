# lint fails for any license not in allowed
allow = ["MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "MPL-2.0", "ISC"]
# any license not explicitly allowed/denied will probably need a legal review (then add to the appropriate list)
deny  = ["GPL-1.0", "GPL-2.0+", "GPL-3.0+",
         "GPL-1.0-only", "GPL-1.0-or-later", "GPL-2.0-only", "GPL-2.0-or-later", "GPL-3.0-only", "GPL-3.0-or-later",
         "AGPL-1.0-only", "AGPL-1.0-or-later", "AGPL-3.0-only", "AGPL-3.0-or-later"]

override = {
  "github.com/gogo/protobuf" = "BSD-3-Clause",
  "github.com/ghodss/yaml" = "MIT",
  "github.com/rcrowley/go-metrics" = "BSD-2-Clause",
  "github.com/magiconair/properties" = "BSD-2-Clause",

  // These aren't true (yet) but they prevent us from erroring out on proprietary bits for now
  "github.com/confluentinc/ccloud-sdk-go" = "Apache-2.0",
  "github.com/confluentinc/ccloudapis" = "Apache-2.0",
  "github.com/confluentinc/protoc-gen-ccloud" = "Apache-2.0",
  "github.com/confluentinc/go-printer" = "Apache-2.0",
  "github.com/confluentinc/go-editor" = "Apache-2.0",
}
