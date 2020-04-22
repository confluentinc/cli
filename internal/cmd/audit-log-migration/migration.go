package audit_log_migration

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
  "sort"

	"github.com/confluentinc/mds-sdk-go"
)

func AuditLogConfigTranslation(clusterConfigs map[string]string, bootstrapServers, crnAuthority string) (mds.AuditLogConfigSpec, error) {
	var newSpec mds.AuditLogConfigSpec
	const defaultTopicName = "confluent-audit-log-events"

	clusterAuditLogConfigSpecs, err := jsonConfigsToAuditLogConfigSpecs(clusterConfigs)
  if err != nil {
    return mds.AuditLogConfigSpec{}, err
  }

  warnMultipleCrnAuthorities(clusterAuditLogConfigSpecs)

  warnMismatchKafaClusters(clusterAuditLogConfigSpecs)

  warnNewBootstrapServers(clusterAuditLogConfigSpecs, bootstrapServers)

	addBootstrapServers(&newSpec, bootstrapServers)

	combineDestinationTopics(clusterAuditLogConfigSpecs, &newSpec)

	setDefaultTopic(&newSpec, defaultTopicName)

	combineExcludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)

  warnNewExcludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)

	combineRoutes(clusterAuditLogConfigSpecs, &newSpec)

	replaceCRNAuthorityRoutes(&newSpec, crnAuthority)

	generateAlternateDefaultTopicRoutes(clusterAuditLogConfigSpecs, &newSpec, crnAuthority)

	return newSpec, nil
}

func warnMultipleCrnAuthorities(specs map[string]*mds.AuditLogConfigSpec) {
  for clusterId, spec := range specs {
    routes := spec.Routes
    foundAuthorities := []string{}
    for routeName, _ := range routes {
      foundAuthority := getCrnAuthority(routeName)
      foundAuthorities = append(foundAuthorities, foundAuthority)
    }
    foundAuthorities = removeDuplicates(foundAuthorities)
    if len(foundAuthorities) != 1 {
      fmt.Printf("Cluster %q had multiple CRN Authorities in its routes: %v.\n", clusterId, foundAuthorities)
    }
  }
}

func getCrnAuthority(routeName string) string {
  re := regexp.MustCompile("crn://[^/]+/")
  return string(re.Find([]byte(routeName)))
}

func warnMismatchKafaClusters(specs map[string]*mds.AuditLogConfigSpec) {
  for clusterId, spec := range specs {
    routes := spec.Routes
    for routeName, _ := range routes {
      if checkMismatchKafkaCluster(routeName, clusterId) {
        fmt.Printf("Cluster %q has a route with a different clusterId. Offending route: %q.\n", clusterId, routeName)
      }
    }
  }
}

func checkMismatchKafkaCluster(routeName, expectedClusterId string) bool {
  re := regexp.MustCompile("(kafka=\\*|kafka="+ expectedClusterId +")")
  result := string(re.Find([]byte(routeName)))
  return result == ""
}

func warnNewBootstrapServers(specs map[string]*mds.AuditLogConfigSpec, bootstrapServers string) {
  for clusterId, spec := range specs {
    old_bootstrap := spec.Destinations.BootstrapServers[0] // assuming there is only a single server

    if old_bootstrap != bootstrapServers {
      fmt.Printf("Cluster %q currently has bootstrap server %q. Replacing with %q.\n", clusterId, old_bootstrap, bootstrapServers)
    }
  }
}

func jsonConfigsToAuditLogConfigSpecs(clusterConfigs map[string]string) (map[string]*mds.AuditLogConfigSpec, error) {
	clusterAuditLogConfigSpecs := make(map[string]*mds.AuditLogConfigSpec)
	for k, v := range clusterConfigs {
		var spec mds.AuditLogConfigSpec
		err := json.Unmarshal([]byte(v), &spec)
    if err !=  nil {
      return nil, err
    }
		clusterAuditLogConfigSpecs[k] = &spec
	}
	return clusterAuditLogConfigSpecs, nil
}

func addBootstrapServers(spec *mds.AuditLogConfigSpec, bootstrapServers string) {
	spec.Destinations.BootstrapServers = []string{bootstrapServers}
}

func combineDestinationTopics(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
	newTopics := make(map[string]mds.AuditLogConfigDestinationConfig)
  topicRetentionDiscrepancies := make(map[string][]int64)

	for _, spec := range specs {
		topics := spec.Destinations.Topics
		for topicName, destination := range topics {
			if _, ok := newTopics[topicName]; ok {
        if destination.RetentionMs != newTopics[topicName].RetentionMs {
          old_retention := min(destination.RetentionMs, newTopics[topicName].RetentionMs)
          handleTopicRetentionDiscrepancy(topicRetentionDiscrepancies, topicName, old_retention)
        }
        newTopics[topicName] = mds.AuditLogConfigDestinationConfig{
          max(destination.RetentionMs, newTopics[topicName].RetentionMs),
        }
			} else {
				newTopics[topicName] = destination
			}
		}
	}

  warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies)

	newSpec.Destinations.Topics = newTopics
}

func handleTopicRetentionDiscrepancy(topicRetentionDiscrepancies map[string][]int64, topicName string, time int64) {
  topicRetentionDiscrepancies[topicName] = append(topicRetentionDiscrepancies[topicName], time)
}

func warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies map[string][]int64) {
  for topicName, retentionTimes := range topicRetentionDiscrepancies {
    fmt.Printf("Topic %q had discrepancies with retention time. The following retention times were discarded: %v.\n", topicName, retentionTimes)
  }
}

func setDefaultTopic(newSpec *mds.AuditLogConfigSpec, defaultTopicName string) {
	const DEFAULT_RETENTION_MS = 7776000000

	newSpec.DefaultTopics = mds.AuditLogConfigDefaultTopics{
		Allowed: defaultTopicName,
		Denied:  defaultTopicName,
	}

	if _, ok := newSpec.Destinations.Topics[defaultTopicName]; !ok {
		newSpec.Destinations.Topics[defaultTopicName] = mds.AuditLogConfigDestinationConfig{
			DEFAULT_RETENTION_MS,
		}
	}
}

func combineExcludedPrincipals(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
	var newExcludedPrincipals []string

	for _, spec := range specs {
		excludedPrincipals := spec.ExcludedPrincipals
		for _, principal := range excludedPrincipals {
			if !find(newExcludedPrincipals, principal) {
				newExcludedPrincipals = append(newExcludedPrincipals, principal)
			}
		}
	}

  sort.Slice(newExcludedPrincipals, func(i, j int) bool {
      return newExcludedPrincipals[i] < newExcludedPrincipals[j]
  })

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

	for crnPath, routeValue := range routes {
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
	re := regexp.MustCompile("^crn://([^/]*)/")
	return string(re.ReplaceAll([]byte(crnPath), []byte("crn://"+newCrnAuthority+"/")))
}

func replaceClusterId(crnPath, clusterId string) string {
	const kafkaIdentifier = "kafka=*"
	if !strings.Contains(crnPath, kafkaIdentifier) {
		fmt.Printf("%q not present in crnPath %q, cannot insert clusterId.\n", kafkaIdentifier, crnPath)
		return crnPath
	}
	return strings.Replace(crnPath, kafkaIdentifier, "kafka="+clusterId, 1)
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
		if defaultTopic != "confluent-audit-log-events" {
			other := mds.AuditLogConfigRouteCategoryTopics{
				Allowed: &defaultTopic,
				Denied:  &defaultTopic,
			}

			// add OTHER block
			for routeName, route := range newSpec.Routes {
				if strings.Contains(routeName, "kafka="+clusterId) {

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
				}
			}
		}
	}
}

func warnNewExcludedPrincipals(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
  for clusterId, spec := range specs {
    excludedPrincipals := spec.ExcludedPrincipals
    differentPrincipals := []string{}
    for _, principal := range newSpec.ExcludedPrincipals {
      if !find(excludedPrincipals, principal) {
        differentPrincipals = append(differentPrincipals, principal)
      }
    }
    if len(differentPrincipals) != 0 {
      fmt.Printf("Cluster %q will now also exclude the following principals: %v.\n", clusterId, differentPrincipals)
    }
  }
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func min(x, y int64) int64 {
	if x < y {
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

func removeDuplicates(s []string) []string {
  check := make(map[string]int)
  for _, v := range s {
    check[v] = 0
  }
  var noDups []string
  for k, _ := range check {
    noDups = append(noDups, k)
  }
  return noDups
}
