package streamshare

import (
	"testing"

	"github.com/stretchr/testify/require"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func TestGetSubjectsCRNFromSharedResources(t *testing.T) {
	subjectCRN := "crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test-value"
	sharedResources := []cdxv1.CdxV1ProviderSharedResource{
		{
			Resources: &[]string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/cloud-cluster=lkc-123/kafka=lkc-123/topic=topic_test",
				subjectCRN,
			},
		},
	}
	subectsCRN, err := getSubjectsCRNFromSharedResources(sharedResources)
	require.NoError(t, err)
	require.Len(t, subectsCRN, 1)
	require.Equal(t, subjectCRN, subectsCRN[0])
}

func TestAreSubjectsModified(t *testing.T) {
	type test struct {
		newCRNs      []string
		existingCRNs []string
		err          error
	}

	expectedError := errors.New(errors.SubjectsListUnmodifiableErrorMsg)
	tests := []test{
		// length is more
		{
			newCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_2-value",
			},
			existingCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
			},
			err: expectedError,
		},
		// length is less
		{
			newCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
			},
			existingCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_2-value",
			},
			err: expectedError,
		},
		// different subjects
		{
			newCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_2-value",
			},
			existingCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
			},
			err: expectedError,
		},
		// changed order
		{
			newCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_2-value",
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
			},
			existingCRNs: []string{
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_1-value",
				"crn://confluent.cloud/organization=abc123/environment=env-123/schema-registry=lsrc-123/subject=topic_test_2-value",
			},
			err: nil,
		},
	}

	for _, tc := range tests {
		err := areSubjectsModified(tc.newCRNs, tc.existingCRNs)
		if tc.err != nil {
			require.EqualError(t, err, tc.err.Error())
		} else {
			require.NoError(t, err)
		}
	}
}
