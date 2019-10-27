package lint_cli

import (
	"strings"
	"testing"

	"github.com/client9/gospell"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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

func TestFlagKebabCase(t *testing.T) {
	rule := RequireFlagKebabCase

	t.Run("invalid flag name", func(t *testing.T) {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("caCertPath", "", "not a valid kebab-case name")
		err := cmd.Execute()
		require.NoError(t, err)
		err = rule(cmd.Flag("caCertPath"), cmd)
		require.Error(t, err)
	})

	t.Run("valid flag name", func(t *testing.T) {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("ca-cert-path", "", "tis a valid kebab-case name")
		err := cmd.Execute()
		require.NoError(t, err)
		err = rule(cmd.Flag("ca-cert-path"), cmd)
		require.NoError(t, err)
	})
}

func TestFlagUsageRealWords(t *testing.T) {
	req := require.New(t)
	rule := RequireFlagUsageRealWords

	sampleDic := `6
fillet
of
a
fenny
snake
`
	aff := strings.NewReader("")
	dic := strings.NewReader(sampleDic)
	gs, err := gospell.NewGoSpellReader(aff, dic)
	req.NoError(err)

	SetVocab(gs)

	t.Run("gibberish", func(t *testing.T) {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("elvish", "", "Mae govannen.")
		err := cmd.Execute()
		require.NoError(t, err)
		err = rule(cmd.Flag("elvish"), cmd)
		require.Error(t, err)
	})

	t.Run("sophisticated prose", func(t *testing.T) {
		cmd := &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}}
		cmd.Flags().String("sonnet", "", "Fillet of a fenny snake.")
		err := cmd.Execute()
		require.NoError(t, err)
		err = rule(cmd.Flag("sonnet"), cmd)
		require.NoError(t, err)
	})
}
