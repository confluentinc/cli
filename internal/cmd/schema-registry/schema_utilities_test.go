package schemaregistry

import (
	"context"
	"fmt"
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/stretchr/testify/require"
)

func TestRequestSchemaById(t *testing.T) {
	tests := []struct {
		name string
		want *string
	}{
		{
			name: "placeholder",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2, err := RequestSchemaWithId(1, "~", "subject", &srsdk.APIClient{DefaultApi: &srsdk.DefaultApiService{}}, context.Background())
			require.NoError(t, err)
			fmt.Println("got1:", got1)
			fmt.Println("got2:", got2)
		})
	}
}
