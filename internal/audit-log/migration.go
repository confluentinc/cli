package auditlog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/imdario/mergo"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/types"
)

func configTranslation(clusterConfigs map[string]string, bootstrapServers []string, crnAuthority string) (mdsv1.AuditLogConfigSpec, []string, error) {
	var newSpec mdsv1.AuditLogConfigSpec
	const defaultTopicName = "confluent-audit-log-events"
	var warnings []string
	var newWarnings []string

	sort.Strings(bootstrapServers)

	clusterAuditLogConfigSpecs, err := jsonConfigsToAuditLogConfigSpecs(clusterConfigs)
	if err != nil {
		return mdsv1.AuditLogConfigSpec{}, warnings, err
	}

	newWarnings = migrateOtherCategoryToManagement(clusterAuditLogConfigSpecs)
	warnings = append(warnings, newWarnings...)

	addDefaultEnabledCategories(clusterAuditLogConfigSpecs, defaultTopicName)

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

	replaceCRNAuthorityRoutes(&newSpec, crnAuthority)

	generateAlternateDefaultTopicRoutes(clusterAuditLogConfigSpecs, &newSpec, crnAuthority)

	sort.Strings(warnings)
	return newSpec, warnings, nil
}

func migrateOtherCategoryToManagement(specs map[string]*mdsv1.AuditLogConfigSpec) []string {
	var warnings []string
	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}
		for routeName, route := range *routes {
			if route.Other != nil {
				if route.Management == nil {
					route.Management = route.Other
					route.Other = nil
				} else if reflect.DeepEqual(route.Management, route.Other) {
					route.Other = nil
				} else {
					warning := fmt.Sprintf(errors.OtherCategoryWarning, routeName, clusterId)
					warnings = append(warnings, warning)
					route.Other = nil
				}
				(*routes)[routeName] = route
			}
		}
	}
	return warnings
}

// Add the AUTHORIZE and MANAGEMENT categories to the route when the default topic is different than the default
// ("confluent-audit-log-events")
func addDefaultEnabledCategories(specs map[string]*mdsv1.AuditLogConfigSpec, defaultTopicName string) {
	for _, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}
		if spec.DefaultTopics.Denied != defaultTopicName || spec.DefaultTopics.Allowed != defaultTopicName {
			enabledCategoryTopics := mdsv1.AuditLogConfigRouteCategoryTopics{
				Allowed: &spec.DefaultTopics.Allowed,
				Denied:  &spec.DefaultTopics.Denied,
			}
			for routeName, route := range *routes {
				if route.Management == nil {
					route.Management = &enabledCategoryTopics
				}
				if route.Authorize == nil {
					route.Authorize = &enabledCategoryTopics
				}
				(*routes)[routeName] = route
			}
		}
	}
}

func warnMultipleCRNAuthorities(specs map[string]*mdsv1.AuditLogConfigSpec) []string {
	var warnings []string
	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}

		var foundAuthorities []string
		for routeName := range *routes {
			foundAuthority := getCRNAuthority(routeName)
			foundAuthorities = append(foundAuthorities, foundAuthority)
		}
		foundAuthorities = types.RemoveDuplicates(foundAuthorities)

		if len(foundAuthorities) > 1 {
			sort.Strings(foundAuthorities)
			newWarning := fmt.Sprintf(errors.MultipleCRNWarning, clusterId, foundAuthorities)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}

func getCRNAuthority(routeName string) string {
	re := regexp.MustCompile("^crn://[^/]+/")
	return re.FindString(routeName)
}

func warnMismatchKafaClusters(specs map[string]*mdsv1.AuditLogConfigSpec) []string {
	var warnings []string
	for clusterId, spec := range specs {
		routes := spec.Routes
		if routes == nil {
			continue
		}
		for routeName := range *routes {
			if checkMismatchKafkaCluster(routeName, clusterId) {
				newWarning := fmt.Sprintf(errors.MismatchedKafkaClusterWarning, clusterId, routeName)
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

func warnNewBootstrapServers(specs map[string]*mdsv1.AuditLogConfigSpec, bootstrapServers []string) []string {
	var warnings []string
	for clusterId, spec := range specs {
		oldBootStrapServers := spec.Destinations.BootstrapServers
		sort.Strings(oldBootStrapServers)
		if !slices.Equal(oldBootStrapServers, bootstrapServers) {
			newWarning := fmt.Sprintf(errors.NewBootstrapWarning, clusterId, oldBootStrapServers, bootstrapServers)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}

func jsonConfigsToAuditLogConfigSpecs(clusterConfigs map[string]string) (map[string]*mdsv1.AuditLogConfigSpec, error) {
	clusterAuditLogConfigSpecs := make(map[string]*mdsv1.AuditLogConfigSpec)
	for clusterId, auditConfig := range clusterConfigs {
		var spec mdsv1.AuditLogConfigSpec
		if err := json.Unmarshal([]byte(auditConfig), &spec); err != nil {
			return nil, fmt.Errorf(`bad input file: the audit log configuration for cluster "%s" uses invalid JSON: %w`, clusterId, err)
		}
		clusterAuditLogConfigSpecs[clusterId] = &spec
	}
	return clusterAuditLogConfigSpecs, nil
}

func addBootstrapServers(spec *mdsv1.AuditLogConfigSpec, bootstrapServers []string) {
	spec.Destinations.BootstrapServers = bootstrapServers
}

func combineDestinationTopics(specs map[string]*mdsv1.AuditLogConfigSpec, newSpec *mdsv1.AuditLogConfigSpec) []string {
	newTopics := make(map[string]mdsv1.AuditLogConfigDestinationConfig)
	topicRetentionDiscrepancies := make(map[string]int64)

	for _, spec := range specs {
		topics := spec.Destinations.Topics
		for topicName, destination := range topics {
			if _, ok := newTopics[topicName]; ok {
				retentionTime := max(destination.RetentionMs, newTopics[topicName].RetentionMs)
				if destination.RetentionMs != newTopics[topicName].RetentionMs {
					topicRetentionDiscrepancies[topicName] = retentionTime
				}
				newTopics[topicName] = mdsv1.AuditLogConfigDestinationConfig{RetentionMs: retentionTime}
			} else {
				newTopics[topicName] = destination
			}
		}
	}

	newSpec.Destinations.Topics = newTopics

	return warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies)
}

func warnTopicRetentionDiscrepancies(topicRetentionDiscrepancies map[string]int64) []string {
	warnings := make([]string, len(topicRetentionDiscrepancies))

	i := 0
	for topicName, maxRetentionTime := range topicRetentionDiscrepancies {
		warnings[i] = fmt.Sprintf(errors.RetentionTimeDiscrepancyWarning, topicName, maxRetentionTime)
		i++
	}

	return warnings
}

func setDefaultTopic(newSpec *mdsv1.AuditLogConfigSpec, defaultTopicName string) {
	const DefaultRetentionMs = 7776000000

	newSpec.DefaultTopics = mdsv1.AuditLogConfigDefaultTopics{
		Allowed: defaultTopicName,
		Denied:  defaultTopicName,
	}

	if _, ok := newSpec.Destinations.Topics[defaultTopicName]; !ok {
		newSpec.Destinations.Topics[defaultTopicName] = mdsv1.AuditLogConfigDestinationConfig{RetentionMs: DefaultRetentionMs}
	}
}

func combineExcludedPrincipals(specs map[string]*mdsv1.AuditLogConfigSpec, newSpec *mdsv1.AuditLogConfigSpec) {
	var newExcludedPrincipals []string

	for _, spec := range specs {
		excludedPrincipals := spec.ExcludedPrincipals
		if excludedPrincipals == nil {
			continue
		}

		for _, principal := range *excludedPrincipals {
			if !slices.Contains(newExcludedPrincipals, principal) {
				newExcludedPrincipals = append(newExcludedPrincipals, principal)
			}
		}
	}

	sort.Strings(newExcludedPrincipals)

	newSpec.ExcludedPrincipals = &newExcludedPrincipals
}

func combineRoutes(specs map[string]*mdsv1.AuditLogConfigSpec, newSpec *mdsv1.AuditLogConfigSpec) []string {
	newRoutes := make(map[string]mdsv1.AuditLogConfigRouteCategories)
	var warnings []string

	clusterIds := make([]string, 0)
	for clusterId := range specs {
		clusterIds = append(clusterIds, clusterId)
	}
	sort.Strings(clusterIds)
	for _, clusterId := range clusterIds {
		spec := specs[clusterId]
		routes := spec.Routes
		if routes == nil {
			continue
		}
		for crnPath, route := range *routes {
			newCRNPath := replaceClusterId(crnPath, clusterId)
			if _, ok := newRoutes[newCRNPath]; ok {
				newWarning := fmt.Sprintf(errors.RepeatedRouteWarning, newCRNPath)
				warnings = append(warnings, newWarning)
			} else {
				newRoutes[newCRNPath] = route
			}
		}
	}

	newSpec.Routes = &newRoutes
	return warnings
}

func replaceCRNAuthorityRoutes(newSpec *mdsv1.AuditLogConfigSpec, newCRNAuthority string) {
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

func generateCRNPath(clusterId, crnAuthority, pathExtension, subcluster string) string {
	path := "crn://" + crnAuthority + "/kafka=" + clusterId
	if subcluster != "" {
		path += "/" + subcluster + "=*"
	}
	if pathExtension != "" {
		path += "/" + pathExtension + "=*"
	}
	return path
}

// For each of the input clusters, we need to make sure that if they specify a default topic different than the global one,
// that messages go to their specific default topics instead of the global default topic
func generateAlternateDefaultTopicRoutes(specs map[string]*mdsv1.AuditLogConfigSpec, newSpec *mdsv1.AuditLogConfigSpec, newCRNAuthority string) {
	type Resource struct {
		extension  string
		categories []string
	}

	type Subcluster struct {
		name      string
		resources []Resource
	}

	// We'll have to add all these routes for each input file
	subclusterRoutes := []Subcluster{
		{
			name: "",
			resources: []Resource{
				{extension: "", categories: []string{"Authorize", "Management"}},
				{extension: "topic", categories: []string{"Authorize", "Management"}},
				{extension: "transaction-id", categories: []string{"Authorize"}},
				{extension: "group", categories: []string{"Authorize", "Management"}},
				{extension: "delegation-token", categories: []string{"Authorize"}},
				{extension: "control-center-broker-metrics", categories: []string{"Authorize"}},
				{extension: "control-center-alerts", categories: []string{"Authorize"}},
				{extension: "cluster-registry", categories: []string{"Authorize"}},
				{extension: "security-metadata", categories: []string{"Authorize"}},
				{extension: "all", categories: []string{"Authorize"}},
			},
		},
		{
			name: "connect",
			resources: []Resource{
				{extension: "", categories: []string{"Authorize"}},
				{extension: "connector", categories: []string{"Authorize"}},
				{extension: "secret", categories: []string{"Authorize"}},
				{extension: "all", categories: []string{"Authorize"}},
			},
		},
		{
			name: "schema-registry",
			resources: []Resource{
				{extension: "", categories: []string{"Authorize"}},
				{extension: "subject", categories: []string{"Authorize"}},
				{extension: "all", categories: []string{"Authorize"}},
			},
		},
		{
			name: "ksql",
			resources: []Resource{
				{extension: "", categories: []string{"Authorize"}},
				{extension: "ksql-cluster", categories: []string{"Authorize"}},
				{extension: "all", categories: []string{"Authorize"}},
			},
		},
	}

	clusterIds := make([]string, 0)
	for clusterId := range specs {
		clusterIds = append(clusterIds, clusterId)
	}
	sort.Strings(clusterIds)
	for _, clusterId := range clusterIds {
		spec := specs[clusterId]
		if spec.DefaultTopics.Denied != newSpec.DefaultTopics.Denied || spec.DefaultTopics.Allowed != newSpec.DefaultTopics.Allowed {
			oldDefaultTopics := mdsv1.AuditLogConfigRouteCategoryTopics{
				Allowed: &spec.DefaultTopics.Allowed,
				Denied:  &spec.DefaultTopics.Denied,
			}

			// add the new routes defined above
			for _, subcluster := range subclusterRoutes {
				for _, resource := range subcluster.resources {
					routeName := generateCRNPath(clusterId, newCRNAuthority, resource.extension, subcluster.name)

					// Create a map of field name to default topics
					categoriesToRoutes := map[string]any{}
					for _, category := range resource.categories {
						categoriesToRoutes[category] = &oldDefaultTopics
					}

					// Initialize our newRouteConfig with values
					newRouteConfig := mdsv1.AuditLogConfigRouteCategories{}
					if err := mergo.Map(&newRouteConfig, categoriesToRoutes); err != nil {
						continue
					}

					newSpecRoutes := *newSpec.Routes
					if _, ok := newSpecRoutes[routeName]; ok {
						// Route already exists in newSpec, so merge it with our new route config
						if err := mergo.Merge(&newRouteConfig, newSpecRoutes[routeName], mergo.WithOverride); err != nil {
							continue
						}
					}
					newSpecRoutes[routeName] = newRouteConfig
				}
			}
		}
	}
}

func warnNewExcludedPrincipals(specs map[string]*mdsv1.AuditLogConfigSpec, newSpec *mdsv1.AuditLogConfigSpec) []string {
	var warnings []string
	for clusterId, spec := range specs {
		excludedPrincipals := spec.ExcludedPrincipals
		if excludedPrincipals == nil {
			continue
		}

		var differentPrincipals []string
		newSpecPrincipals := *newSpec.ExcludedPrincipals
		for _, principal := range newSpecPrincipals {
			if !slices.Contains(*excludedPrincipals, principal) {
				differentPrincipals = append(differentPrincipals, principal)
			}
		}
		if len(differentPrincipals) != 0 {
			newWarning := fmt.Sprintf(errors.NewExcludedPrincipalsWarning, clusterId, differentPrincipals)
			warnings = append(warnings, newWarning)
		}
	}
	return warnings
}
