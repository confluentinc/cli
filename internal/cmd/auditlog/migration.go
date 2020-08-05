package auditlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/utils"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

func AuditLogConfigTranslation(clusterConfigs map[string]string, bootstrapServers []string, crnAuthority string) (mds.AuditLogConfigSpec, []string, error) {
	var newSpec mds.AuditLogConfigSpec
	const defaultTopicName = "confluent-audit-log-events"
	warnings := []string{}
	var newWarnings []string

	sort.Strings(bootstrapServers)

	clusterAuditLogConfigSpecs, err := jsonConfigsToAuditLogConfigSpecs(clusterConfigs)
	if err != nil {
		return mds.AuditLogConfigSpec{}, warnings, err
	}

	addOtherBlock(clusterAuditLogConfigSpecs, defaultTopicName)

	newWarnings = warnMultipleCRNAuthorities(clusterAuditLogConfigSpecs)
	warnings = append(warnings, newWarnings...)

	newWarnings = warnMismatchKafaClusters(clusterAuditLogConfigSpecs)
	warnings = append(warnings, newWarnings...)

	newWarnings = warnNewBootstrapServers(clusterAuditLogConfigSpecs, bootstrapServers)
	warnings = append(warnings, newWarnings...)

	addBootstrapServers(&newSpec, bootstrapServers)

	newWarnings = combineDestinationTopics(clusterAuditLogConfigSpecs, &newSpec)
	warnings = append(warnings, newWarnings...)

	setDefaultTopic(&newSpec, defaultTopicName)

	combineExcludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)

	newWarnings = warnNewExcludedPrincipals(clusterAuditLogConfigSpecs, &newSpec)
	warnings = append(warnings, newWarnings...)

	newWarnings = combineRoutes(clusterAuditLogConfigSpecs, &newSpec)
	warnings = append(warnings, newWarnings...)

	generateAlternateDefaultTopicRoutes(clusterAuditLogConfigSpecs, &newSpec, crnAuthority)

	replaceCRNAuthorityRoutes(&newSpec, crnAuthority)

	sort.Strings(warnings)

	return newSpec, warnings, nil
}

// add the OTHER block to the route when the default topic is different than the default ("confluent-audit-log-events")
func addOtherBlock(specs map[string]*mds.AuditLogConfigSpec, defaultTopicName string) {
	for _, spec := range specs {
		if spec.DefaultTopics.Denied != defaultTopicName || spec.DefaultTopics.Allowed != defaultTopicName {
			other := mds.AuditLogConfigRouteCategoryTopics{
				Allowed: &spec.DefaultTopics.Allowed,
				Denied:  &spec.DefaultTopics.Denied,
			}
			routes := spec.Routes
			if routes == nil {
				continue
			}

			for routeName, route := range *routes {
				if route.Other == nil {
					route.Other = &other
					(*routes)[routeName] = route
				}
			}
		}
	}
}

func warnMultipleCRNAuthorities(specs map[string]*mds.AuditLogConfigSpec) []string {
	warnings := []string{}
	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}

		foundAuthorities := []string{}
		for routeName := range *routes {
			foundAuthority := getCRNAuthority(routeName)
			foundAuthorities = append(foundAuthorities, foundAuthority)
		}
		foundAuthorities = utils.RemoveDuplicates(foundAuthorities)

		if len(foundAuthorities) > 1 {
			sort.Strings(foundAuthorities)
			newWarning := fmt.Sprintf("Multiple CRN Authorities Warning: Cluster %q had multiple CRN Authorities in its routes: %v.", clusterId, foundAuthorities)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}

func getCRNAuthority(routeName string) string {
	re := regexp.MustCompile("^crn://[^/]+/")
	return re.FindString(routeName)
}

func warnMismatchKafaClusters(specs map[string]*mds.AuditLogConfigSpec) []string {
	warnings := []string{}
	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}
		for routeName := range *routes {
			if checkMismatchKafkaCluster(routeName, clusterId) {
				newWarning := fmt.Sprintf("Mismatched Kafka Cluster Warning: Cluster %q has a route with a different clusterId. Route: %q.", clusterId, routeName)
				warnings = append(warnings, newWarning)
			}
		}
	}
	return warnings
}

func checkMismatchKafkaCluster(routeName, expectedClusterId string) bool {
	re := regexp.MustCompile("/kafka=(\\*|" + regexp.QuoteMeta(expectedClusterId) + ")(?:$|/)")
	result := re.FindString(routeName)
	return result == ""
}

func warnNewBootstrapServers(specs map[string]*mds.AuditLogConfigSpec, bootstrapServers []string) []string {
	warnings := []string{}
	for clusterId, spec := range specs {
		oldBootStrapServers := spec.Destinations.BootstrapServers
		sort.Strings(oldBootStrapServers)
		if !utils.TestEq(oldBootStrapServers, bootstrapServers) {
			newWarning := fmt.Sprintf("New Bootstrap Servers Warning: Cluster %q currently has bootstrap servers = %v. Replacing with %v.", clusterId, oldBootStrapServers, bootstrapServers)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}

func jsonConfigsToAuditLogConfigSpecs(clusterConfigs map[string]string) (map[string]*mds.AuditLogConfigSpec, error) {
	clusterAuditLogConfigSpecs := make(map[string]*mds.AuditLogConfigSpec)
	for k, v := range clusterConfigs {
		var spec mds.AuditLogConfigSpec
		err := json.Unmarshal([]byte(v), &spec)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cluster '%s' has a malformed audit log configuration: %s", k, err.Error()))
		}
		clusterAuditLogConfigSpecs[k] = &spec
	}
	return clusterAuditLogConfigSpecs, nil
}

func addBootstrapServers(spec *mds.AuditLogConfigSpec, bootstrapServers []string) {
	spec.Destinations.BootstrapServers = bootstrapServers
}

func combineDestinationTopics(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) []string {
	newTopics := make(map[string]mds.AuditLogConfigDestinationConfig)
	topicRetentionDiscrepancies := make(map[string]int64)

	for _, spec := range specs {
		topics := spec.Destinations.Topics
		for topicName, destination := range topics {
			if _, ok := newTopics[topicName]; ok {
				retentionTime := utils.Max(destination.RetentionMs, newTopics[topicName].RetentionMs)
				if destination.RetentionMs != newTopics[topicName].RetentionMs {
					topicRetentionDiscrepancies[topicName] = retentionTime
				}
				newTopics[topicName] = mds.AuditLogConfigDestinationConfig{
					RetentionMs: retentionTime,
				}
			} else {
				newTopics[topicName] = destination
			}
		}
	}

	newSpec.Destinations.Topics = newTopics

	return warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies)
}

func warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies map[string]int64) []string {
	warnings := []string{}
	for topicName, maxRetentionTime := range topicRetentionDiscrepancies {
		newWarning := fmt.Sprintf("Retention Time Discrepancy Warning: Topic %q had discrepancies with retention time. Using max: %v.", topicName, maxRetentionTime)
		warnings = append(warnings, newWarning)
	}
	return warnings
}

func setDefaultTopic(newSpec *mds.AuditLogConfigSpec, defaultTopicName string) {
	const DEFAULT_RETENTION_MS = 7776000000

	newSpec.DefaultTopics = mds.AuditLogConfigDefaultTopics{
		Allowed: defaultTopicName,
		Denied:  defaultTopicName,
	}

	if _, ok := newSpec.Destinations.Topics[defaultTopicName]; !ok {
		newSpec.Destinations.Topics[defaultTopicName] = mds.AuditLogConfigDestinationConfig{
			RetentionMs: DEFAULT_RETENTION_MS,
		}
	}
}

func combineExcludedPrincipals(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) {
	var newExcludedPrincipals []string

	for _, spec := range specs {
		excludedPrincipals := spec.ExcludedPrincipals
		if excludedPrincipals == nil {
			continue;
		}

		for _, principal := range *excludedPrincipals {
			if !utils.Contains(newExcludedPrincipals, principal) {
				newExcludedPrincipals = append(newExcludedPrincipals, principal)
			}
		}
	}

	sort.Strings(newExcludedPrincipals)

	newSpec.ExcludedPrincipals = &newExcludedPrincipals
}

func combineRoutes(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) []string {
	newRoutes := make(map[string]mds.AuditLogConfigRouteCategories)
	warnings := []string{}

	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}
		for crnPath, route := range *routes {
			newCRNPath := replaceClusterId(crnPath, clusterId)
			if _, ok := newRoutes[newCRNPath]; ok {
				newWarning := fmt.Sprintf("Repeated Route Warning: Route Name : %q.", newCRNPath)
				warnings = append(warnings, newWarning)
			} else {
				newRoutes[newCRNPath] = route
			}
		}
	}

	newSpec.Routes = &newRoutes
	return warnings
}

func replaceCRNAuthorityRoutes(newSpec *mds.AuditLogConfigSpec, newCRNAuthority string) {
	routes := *newSpec.Routes

	for crnPath, routeValue := range routes {
		if !crnPathContainsAuthority(crnPath, newCRNAuthority) {
			newCRNPath := replaceCRNAuthority(crnPath, newCRNAuthority)
			routes[newCRNPath] = routeValue
			delete(routes, crnPath)
		}
	}
}

func crnPathContainsAuthority(crnPath, crnAuthority string) bool {
	re := regexp.MustCompile("^crn://" + crnAuthority + "/.*")
	return re.MatchString(crnPath)
}

func replaceCRNAuthority(crnPath, newCRNAuthority string) string {
	re := regexp.MustCompile("^crn://([^/]*)/")
	return re.ReplaceAllString(crnPath, "crn://"+newCRNAuthority+"/")
}

func replaceClusterId(crnPath, clusterId string) string {
	const kafkaIdentifier = "kafka=*"
	if !strings.Contains(crnPath, kafkaIdentifier) {
		// crnPath already has a specific kafka cluster, no need to insert clusterId
		return crnPath
	}
	return strings.Replace(crnPath, kafkaIdentifier, "kafka="+clusterId, 1)
}

func generateCRNPath(clusterId, crnAuthority, pathExtension string) string {
	path := "crn://" + crnAuthority + "/kafka=" + clusterId
	if pathExtension != "" {
		path += "/" + pathExtension + "=*"
	}
	return path
}

func generateAlternateDefaultTopicRoutes(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec, crnAuthority string) {
	for clusterId, spec := range specs {
		if spec.DefaultTopics.Denied != newSpec.DefaultTopics.Denied || spec.DefaultTopics.Allowed != newSpec.DefaultTopics.Allowed {
			other := mds.AuditLogConfigRouteCategoryTopics{
				Allowed: &spec.DefaultTopics.Allowed,
				Denied:  &spec.DefaultTopics.Denied,
			}

			// add the four new routes to the newSpec, if not already there
			newRouteConfig := mds.AuditLogConfigRouteCategories{
				Other: &other,
			}
			pathExtensions := []string{"", "topic", "connect", "ksql"}
			for _, extension := range pathExtensions {
				routeName := generateCRNPath(clusterId, crnAuthority, extension)
				newSpecRoutes := *newSpec.Routes
				if _, ok := newSpecRoutes[routeName]; !ok {
					newSpecRoutes[routeName] = newRouteConfig
				}
			}
		}
	}
}

func warnNewExcludedPrincipals(specs map[string]*mds.AuditLogConfigSpec, newSpec *mds.AuditLogConfigSpec) []string {
	warnings := []string{}
	for clusterId, spec := range specs {
		excludedPrincipals := spec.ExcludedPrincipals
		if excludedPrincipals == nil {
			continue
		}

		differentPrincipals := []string{}
		newSpecPrincipals := *newSpec.ExcludedPrincipals
		for _, principal := range newSpecPrincipals {
			if !utils.Contains(*excludedPrincipals, principal) {
				differentPrincipals = append(differentPrincipals, principal)
			}
		}
		if len(differentPrincipals) != 0 {
			newWarning := fmt.Sprintf("New Excluded Principals Warning: Cluster %q will now also exclude the following principals: %v.", clusterId, differentPrincipals)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}
