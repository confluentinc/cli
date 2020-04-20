package audit_log_migration

import (
  "fmt"
  "encoding/json"

  "github.com/confluentinc/mds-sdk-go"
)

func AuditLogConfigTranslation(clusterConfigs map[string]string, bootstrapServers, crnAuthority string) mds.AuditLogConfigSpec {

  var newSpec mds.AuditLogConfigSpec

  //step 0 - turn clusterConfigs into a map of <string>*specs
  clusterAuditLogConfigSpecs := jsonConfigsToAuditLogConfigSpecs(clusterConfigs)

  // step 1 - add bootstrapServers
  addBootstrapServers(&newSpec, bootstrapServers)

  // step 2 - combine destination topics
  combineDestinationTopics(clusterAuditLogConfigSpecs, &newSpec)

  // step 3 - set default
  setDefaultTopic(&newSpec, "confluent-audit-log-events") // TODO: make this a CONST somehwere

  // step 4 - combine exlcuded ExcludedPrincipals
  //combineExludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)



  printSpec(&newSpec)

  fmt.Println(crnAuthority)

  return newSpec
}

func printSpec(spec *mds.AuditLogConfigSpec) {
  fmt.Println(spec.Destinations)
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

func addBootstrapServers(spec *mds.AuditLogConfigSpec, bootstrapServers string) {
  spec.Destinations.BootstrapServers = []string{bootstrapServers}
}

func combineDestinationTopics(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
  newTopics := make(map[string]mds.AuditLogConfigDestinationConfig)

  // lazy merge
  for _, spec := range specs {
    topics := spec.Destinations.Topics
    for k, destination := range topics {
      if _, ok := newTopics[k]; ok {
        newTopics[k] = mds.AuditLogConfigDestinationConfig{
          max(destination.RetentionMs, newTopics[k].RetentionMs),
        }
      } else {
        newTopics[k] = destination
      }
    }
  }

  newSpec.Destinations.Topics = newTopics
}

func setDefaultTopic(newSpec *mds.AuditLogConfigSpec, defaultTopicName string) {
  newSpec.DefaultTopics = mds.AuditLogConfigDefaultTopics{
    Allowed: defaultTopicName,
    Denied: defaultTopicName,
  }
  if _, ok := newSpec.Destinations.Topics[defaultTopicName]; !ok {
    newSpec.Destinations.Topics[defaultTopicName] = mds.AuditLogConfigDestinationConfig{
      7776000000,
    }
  }
}

// func combineExludedPrincipals(newSpec *mds.AuditLogConfigSpec, defaultTopicName string) {
//
// }

func max(x, y int64) int64 {
 if x > y {
   return x
 }
 return y
}
