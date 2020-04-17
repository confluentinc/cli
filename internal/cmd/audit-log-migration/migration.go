package audit_log_migration

import (
  "fmt"
  "encoding/json"

  "github.com/confluentinc/mds-sdk-go"
)

func AuditLogConfigTranslation(clusterConfigs map[string]string, bootstrapServers, crnAuthority string) mds.AuditLogConfigSpec {

  //step 0 - turn clusterConfigs into a map of <string>*specs
  clusterAuditLogConfigSpecs := jsonConfigsToAuditLogConfigSpecs(clusterConfigs)

  // step 1 - add bootstrapServers
  addBootstrapServers(clusterAuditLogConfigSpecs, bootstrapServers)

  // step 2 -


  for _,v := range clusterAuditLogConfigSpecs {
    printSpec(v)
  }

  fmt.Println(crnAuthority)

  return *clusterAuditLogConfigSpecs["cluster123"]
}

func printSpec(spec *mds.AuditLogConfigSpec) {
  fmt.Println(spec.Routes)
}

func jsonConfigsToAuditLogConfigSpecs(clusterConfigs map[string]string) map[string]*mds.AuditLogConfigSpec {
  clusterAuditLogConfigSpecs := make(map[string]*mds.AuditLogConfigSpec)
  for k, v := range clusterConfigs {
    var spec mds.AuditLogConfigSpec
    json.Unmarshal([]byte(v), &spec)
    clusterAuditLogConfigSpecs[k] = &spec
  }
  return clusterAuditLogConfigSpecs
}

func addBootstrapServers(specs map[string]*mds.AuditLogConfigSpec, bootstrapServers string) {
  for k, _ := range specs {
    specs[k].Destinations.BootstrapServers = []string{bootstrapServers}
  }
}
