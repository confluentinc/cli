package audit_log_migration

import (
  "fmt"
  "encoding/json"
  "regexp"
  "strings"

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
  combineExludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)

  // step 5 - combine routes and Replace the CRN authority
  combineRoutes(clusterAuditLogConfigSpecs, &newSpec)
  replaceCRNAuthorityRoutes(&newSpec, crnAuthority)

  // step 6 - generate routes for default topics
  generateAlternateDefaultTopicRoutes(clusterAuditLogConfigSpecs, &newSpec, crnAuthority)

  printSpec(&newSpec)
  return newSpec
}

func printSpec(spec *mds.AuditLogConfigSpec) {
  for k, v := range spec.Routes {
    fmt.Println(k)
    if v.Other != nil {
      fmt.Println(*v.Other.Allowed)
    }
  }
}

func jsonConfigsToAuditLogConfigSpecs(clusterConfigs map[string]string) map[string]*mds.AuditLogConfigSpec {
  clusterAuditLogConfigSpecs := make(map[string]*mds.AuditLogConfigSpec)
  for k, v := range clusterConfigs {
    var spec mds.AuditLogConfigSpec
    json.Unmarshal([]byte(v), &spec)
    fmt.Println(v)
    fmt.Println(spec)
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

func combineExludedPrincipals(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
  var newExcludedPrincipals []string

  for _, spec := range specs {
    excludedPrincipals := spec.ExcludedPrincipals
    for _, principal := range excludedPrincipals {
      if !find(newExcludedPrincipals, principal) {
        newExcludedPrincipals = append(newExcludedPrincipals, principal)
      }
    }
  }

  newSpec.ExcludedPrincipals = newExcludedPrincipals
}

func combineRoutes(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
  newRoutes := make(map[string]mds.AuditLogConfigRouteCategories)

  for clusterId, spec := range specs {
    routes := spec.Routes
    for crnPath, route := range routes {
      newCrnPath := replaceClusterId(crnPath, clusterId)
      newRoutes[newCrnPath] = route
    }
  }

  newSpec.Routes = newRoutes
}

func replaceCRNAuthorityRoutes(newSpec *mds.AuditLogConfigSpec, newCrnAuthority string) {
  routes := newSpec.Routes
  for k, _ := range newSpec.Routes {
    fmt.Println(k)
  }
  for crnPath, routeValue := range routes {
    // if crn path is already valid, no need to replace!
    if !crnPathContainsAuthority(crnPath, newCrnAuthority) {
      newCrnPath := replaceCRNAuthority(crnPath, newCrnAuthority)
      routes[newCrnPath] = routeValue
      delete(routes, crnPath)
    }
  }
}

func crnPathContainsAuthority(crnPath, crnAuthority string) bool {
  re := regexp.MustCompile("^crn://" + crnAuthority + "/.*")
  return re.Match([]byte(crnPath))
}

func replaceCRNAuthority(crnPath, newCrnAuthority string) string {
  re := regexp.MustCompile("crn://([^/]*)/")
  return string(re.ReplaceAll([]byte(crnPath), []byte("crn://" + newCrnAuthority + "/")))
}

func replaceClusterId(crnPath, clusterId string) string {
  const kafkaIdentifier = "kafka=*"
  if !strings.Contains(crnPath, kafkaIdentifier) {
    err := fmt.Errorf("%q not present in crnPath %q, cannot insert clusterId", kafkaIdentifier, crnPath)
  	fmt.Println(err.Error())
    return crnPath
  }
  return strings.Replace(crnPath, kafkaIdentifier, "kafka=" + clusterId, 1)
}

func generateCrnPath(clusterId, crnAuthority, pathExtension string) string {
  path := "crn://" + crnAuthority + "/kafka=" + clusterId
  if pathExtension != "" {
    path += "/" + pathExtension + "=*"
  }
  return path
}

func generateAlternateDefaultTopicRoutes(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec, crnAuthority string) {
  for clusterId, spec := range specs {
    defaultTopic := spec.DefaultTopics.Denied
    fmt.Println(clusterId)
    if defaultTopic != "confluent-audit-log-events" {
      other := mds.AuditLogConfigRouteCategoryTopics{
        Allowed: &defaultTopic,
        Denied: &defaultTopic,
      }

      // // add OTHER block
      for routeName, route := range newSpec.Routes {
        if strings.Contains(routeName, "kafka=" + clusterId) {

          if newSpec.Routes[routeName].Other == nil {
            route.Other = &other
            newSpec.Routes[routeName] = route
          }
        }
      }

      // add the four new routes to the newSpec, if not already there
      newRouteConfig := mds.AuditLogConfigRouteCategories{
        Other: &other,
      }
      pathExtensions := []string{"", "topic", "connect", "ksql"}
      for _, extension := range pathExtensions {
        routeName := generateCrnPath(clusterId, crnAuthority, extension)
        if _, ok := newSpec.Routes[routeName]; !ok {
          newSpec.Routes[routeName] = newRouteConfig
          fmt.Println("replace")
          fmt.Println(*newSpec.Routes[routeName].Other.Allowed)
        }
      }


    }
  }
}

func max(x, y int64) int64 {
 if x > y {
   return x
 }
 return y
}

func find(slice []string, val string) bool {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}
