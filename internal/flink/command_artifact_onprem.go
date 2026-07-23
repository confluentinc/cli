package flink

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/properties"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

// CMF strictly validates these on create and update; the CLI always sends the standard values.
const (
	artifactApiVersion = "cmf.confluent.io/v1"
	artifactKind       = "Artifact"
)

// artifactLookupSuggestions is shared by the delete handlers when an artifact can't be found (or CMF is unreachable).
const artifactLookupSuggestions = "List available Flink artifacts with `confluent flink artifact list`." +
	"\nCheck that CMF is running and accessible."

// artifactOutOnPrem is the human-readable row used by the artifact list command (kept lean; no labels/annotations).
type artifactOutOnPrem struct {
	Name         string `human:"Name" serialized:"name"`
	Version      string `human:"Version" serialized:"version"`
	Phase        string `human:"Phase" serialized:"phase"`
	Size         string `human:"Size" serialized:"size"`
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
}

// artifactDescribeOutOnPrem is the human-readable view for the single-artifact commands (describe, create, update). It
// adds Labels and Annotations (omitted when empty) so users can see managed metadata without switching to -o json/yaml.
type artifactDescribeOutOnPrem struct {
	Name         string            `human:"Name"`
	Version      string            `human:"Version"`
	Phase        string            `human:"Phase"`
	Size         string            `human:"Size"`
	CreationTime string            `human:"Creation Time"`
	Labels       map[string]string `human:"Labels,omitempty"`
	Annotations  map[string]string `human:"Annotations,omitempty"`
}

// artifactVersionOutOnPrem is the human-readable view of a single artifact version (version create, version list, version describe).
type artifactVersionOutOnPrem struct {
	Version      string `human:"Version" serialized:"version"`
	Phase        string `human:"Phase" serialized:"phase"`
	Size         string `human:"Size" serialized:"size"`
	Checksum     string `human:"Checksum" serialized:"checksum"`
	CreationTime string `human:"Creation Time" serialized:"creation_time"`
}

func newArtifactOutOnPrem(artifact cmfsdk.Artifact) *artifactOutOnPrem {
	out := &artifactOutOnPrem{Name: artifact.Metadata.Name}
	// Artifact-level view: Creation Time is when the artifact was first created (Metadata), not the current version.
	// The version-level view below deliberately uses Status.CreationTimestamp instead; keep the two sources distinct.
	if artifact.Metadata.CreationTimestamp != nil {
		out.CreationTime = *artifact.Metadata.CreationTimestamp
	}
	if artifact.Status != nil {
		if artifact.Status.Version != nil {
			out.Version = strconv.Itoa(int(*artifact.Status.Version))
		}
		if artifact.Status.Phase != nil {
			out.Phase = *artifact.Status.Phase
		}
		if artifact.Status.Size != nil {
			out.Size = strconv.FormatInt(*artifact.Status.Size, 10)
		}
	}
	return out
}

func newArtifactVersionOutOnPrem(artifact cmfsdk.Artifact) *artifactVersionOutOnPrem {
	out := &artifactVersionOutOnPrem{}
	if artifact.Status != nil {
		if artifact.Status.Version != nil {
			out.Version = strconv.Itoa(int(*artifact.Status.Version))
		}
		if artifact.Status.Phase != nil {
			out.Phase = *artifact.Status.Phase
		}
		if artifact.Status.Size != nil {
			out.Size = strconv.FormatInt(*artifact.Status.Size, 10)
		}
		if artifact.Status.Checksum != nil {
			out.Checksum = *artifact.Status.Checksum
		}
		// Version-level view: Creation Time is when this specific version was uploaded (Status), not the artifact.
		// This intentionally differs from newArtifactOutOnPrem, which reads Metadata.CreationTimestamp.
		if artifact.Status.CreationTimestamp != nil {
			out.CreationTime = *artifact.Status.CreationTimestamp
		}
	}
	return out
}

// newArtifactDescribeOutOnPrem builds the single-artifact human view, reusing the list row for the shared columns and
// adding Labels/Annotations. They are left nil (and thus omitted from the table) when the artifact has none.
func newArtifactDescribeOutOnPrem(artifact cmfsdk.Artifact) *artifactDescribeOutOnPrem {
	base := newArtifactOutOnPrem(artifact)
	out := &artifactDescribeOutOnPrem{
		Name:         base.Name,
		Version:      base.Version,
		Phase:        base.Phase,
		Size:         base.Size,
		CreationTime: base.CreationTime,
	}
	// These are value-type maps with `human:"...,omitempty"`, so assigning a nil map is a no-op (the row is omitted).
	// Unlike convertSdkArtifactToLocalArtifact, no nil guard is needed here.
	out.Labels = artifact.Metadata.Labels
	out.Annotations = artifact.Metadata.Annotations
	return out
}

// printArtifactOnPrem prints a single artifact (artifact-level view).
func printArtifactOnPrem(cmd *cobra.Command, artifact cmfsdk.Artifact) error {
	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(newArtifactDescribeOutOnPrem(artifact))
		return table.Print()
	}
	return output.SerializedOutput(cmd, convertSdkArtifactToLocalArtifact(artifact))
}

// printArtifactVersionOnPrem prints a single artifact version (version-level view).
func printArtifactVersionOnPrem(cmd *cobra.Command, artifact cmfsdk.Artifact) error {
	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		table.Add(newArtifactVersionOutOnPrem(artifact))
		return table.Print()
	}
	return output.SerializedOutput(cmd, convertSdkArtifactToLocalArtifact(artifact))
}

func convertSdkArtifactToLocalArtifact(sdkArtifact cmfsdk.Artifact) LocalArtifact {
	localArtifact := LocalArtifact{
		ApiVersion: sdkArtifact.ApiVersion,
		Kind:       sdkArtifact.Kind,
		Metadata: LocalArtifactMetadata{
			Name:              sdkArtifact.Metadata.Name,
			CreationTimestamp: sdkArtifact.Metadata.CreationTimestamp,
			UpdateTimestamp:   sdkArtifact.Metadata.UpdateTimestamp,
			Uid:               sdkArtifact.Metadata.Uid,
		},
		Spec: sdkArtifact.Spec,
	}

	if sdkArtifact.Metadata.Labels != nil {
		localArtifact.Metadata.Labels = &sdkArtifact.Metadata.Labels
	}
	if sdkArtifact.Metadata.Annotations != nil {
		localArtifact.Metadata.Annotations = &sdkArtifact.Metadata.Annotations
	}

	if sdkArtifact.Status != nil {
		localArtifact.Status = &LocalArtifactStatus{
			Version:           sdkArtifact.Status.Version,
			CreationTimestamp: sdkArtifact.Status.CreationTimestamp,
			Path:              sdkArtifact.Status.Path,
			Size:              sdkArtifact.Status.Size,
			Checksum:          sdkArtifact.Status.Checksum,
			Phase:             sdkArtifact.Status.Phase,
			Message:           sdkArtifact.Status.Message,
		}
	}

	return localArtifact
}

// newSdkArtifact builds an Artifact payload for create and update requests. Labels are set only when non-nil so that
// omitting them (nil) preserves any existing labels server-side.
func newSdkArtifact(name string, labels map[string]string) cmfsdk.Artifact {
	artifact := cmfsdk.Artifact{
		ApiVersion: artifactApiVersion,
		Kind:       artifactKind,
		Metadata:   cmfsdk.ArtifactMetadata{Name: name},
		Spec:       map[string]interface{}{},
	}
	if labels != nil {
		artifact.Metadata.Labels = labels
	}
	return artifact
}

// getLabelsFlag parses the repeatable "--label key=value" flag into a map. It returns a nil map when the flag was not provided.
func getLabelsFlag(cmd *cobra.Command) (map[string]string, error) {
	labelSlice, err := cmd.Flags().GetStringSlice("label")
	if err != nil {
		return nil, err
	}
	if len(labelSlice) == 0 {
		return nil, nil
	}
	return properties.ConfigSliceToMap(labelSlice)
}

// openArtifactFile validates the artifact file extension and opens it for upload. The caller is responsible for closing the returned file.
func openArtifactFile(path string) (*os.File, error) {
	extension := strings.TrimPrefix(filepath.Ext(path), ".")
	if !slices.Contains(allowedFileExtensions, strings.ToLower(extension)) {
		return nil, fmt.Errorf("only extensions allowed for `--artifact-file` are %s", utils.ArrayToCommaDelimitedString(allowedFileExtensions, "and"))
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact file: %w", err)
	}

	return file, nil
}
