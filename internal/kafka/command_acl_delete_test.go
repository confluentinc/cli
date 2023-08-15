package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintAclsDeleted(t *testing.T) {
	assert.Equal(t, "ACL not found. ACL may have been misspelled or already deleted.", printAclsDeleted(0))
	assert.Equal(t, "Deleted 1 ACL.", printAclsDeleted(1))
	assert.Equal(t, "Deleted 2 ACLs.", printAclsDeleted(2))
}
