package lint_cli

import (
	"testing"
)

func passes(fieldValue string, properNouns []string, t *testing.T) {
	// "field" and "fullCommand" are only used in error printing format -- immaterial to functionality
	if requireNotTitleCaseHelper(fieldValue, properNouns, "f", "fc").ErrorOrNil() != nil {
		t.Fail()
	}
}

func fails(fieldValue string, properNouns []string, t *testing.T) {
	// "field" and "fullCommand" are only used in error printing format -- immaterial to functionality
	if requireNotTitleCaseHelper(fieldValue, properNouns, "f", "fc").ErrorOrNil() == nil {
		t.Fail()
	}
}

func TestRequireNotTitleCase(t *testing.T) {
	passes("This is fine.", []string{}, t)
	fails("This Isn't fine.", []string{}, t)
	passes("This is fine Kafka.", []string{"Kafka"}, t)
	fails("This is not fine Schema Registry.", []string{"Kafka"}, t)
	passes("This is fine Schema Registry, though.", []string{"Schema Registry"}, t)
	fails("not ok.", []string{"Schema Registry"}, t)
	passes("Connect team. Loves to hack.", []string{}, t)
}
