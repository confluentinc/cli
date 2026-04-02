//go:build live_test && (all || iam)

package live

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// generateSelfSignedCert creates a self-signed CA certificate and returns it as a base64 encoded PEM string.
func generateSelfSignedCert(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"CLI Live Test"},
			CommonName:   "cli-live-test-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return base64.StdEncoding.EncodeToString(certPEM)
}

func (s *CLILiveTestSuite) TestIAMCertificateAuthorityCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	certChain := generateSelfSignedCert(t)

	// Register cleanups in LIFO order: pool first, then authority
	s.registerCleanup(t, "iam certificate-authority delete {{.cert_authority_id}} --force", state)
	s.registerCleanup(t, "iam certificate-pool delete {{.cert_pool_id}} --provider {{.cert_authority_id}} --force", state)

	steps := []CLILiveTest{
		// Certificate Authority CRUD
		{
			Name:      "Create certificate authority",
			Args:      `iam certificate-authority create --description "Live test CA" --certificate-chain "` + certChain + `" --certificate-chain-filename live-test-ca.pem -o json`,
			CaptureID: "cert_authority_id",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:         "Describe certificate authority",
			Args:         "iam certificate-authority describe {{.cert_authority_id}} -o json",
			UseStateVars: true,
			JSONFieldsExist: []string{"id", "description"},
		},
		{
			Name: "List certificate authorities",
			Args: "iam certificate-authority list",
		},
		{
			Name:         "Update certificate authority",
			Args:         `iam certificate-authority update {{.cert_authority_id}} --description "Updated live test CA"`,
			UseStateVars: true,
		},
		// Certificate Pool CRUD (depends on authority)
		{
			Name:         "Create certificate pool",
			Args:         `iam certificate-pool create --provider {{.cert_authority_id}} --display-name "live-test-pool" --description "Live test certificate pool" --external-identifier "OU=Engineering" --filter 'certificate.subject == "OU=Engineering"' -o json`,
			UseStateVars: true,
			CaptureID:    "cert_pool_id",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:         "Describe certificate pool",
			Args:         "iam certificate-pool describe {{.cert_pool_id}} --provider {{.cert_authority_id}} -o json",
			UseStateVars: true,
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:         "List certificate pools",
			Args:         "iam certificate-pool list --provider {{.cert_authority_id}}",
			UseStateVars: true,
		},
		{
			Name:         "Update certificate pool",
			Args:         `iam certificate-pool update {{.cert_pool_id}} --provider {{.cert_authority_id}} --description "Updated live test pool"`,
			UseStateVars: true,
		},
		{
			Name:         "Delete certificate pool",
			Args:         "iam certificate-pool delete {{.cert_pool_id}} --provider {{.cert_authority_id}} --force",
			UseStateVars: true,
		},
		// Clean up certificate authority
		{
			Name:         "Delete certificate authority",
			Args:         "iam certificate-authority delete {{.cert_authority_id}} --force",
			UseStateVars: true,
		},
		{
			Name:         "Verify certificate authority deleted",
			Args:         "iam certificate-authority describe {{.cert_authority_id}}",
			UseStateVars: true,
			ExitCode:     1,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
