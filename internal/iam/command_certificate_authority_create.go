package iam

import (
	"github.com/spf13/cobra"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *certificateAuthorityCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a certificate authority.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create the certificate authority "my-ca" using the certificate chain stored in the "CERTIFICATE_CHAIN" environment variable:`,
				Code: `confluent iam certificate-authority create my-ca --description "my certificate authority" --certificate-chain $CERTIFICATE_CHAIN --certificate-chain-filename certificate.pem`,
			},
			examples.Example{
				Text: "An example of a certificate chain:",
				Code: `-----BEGIN CERTIFICATE-----
MIIDdTCCAl2gAwIBAgILBAAAAAABFUtaw5QwDQYJKoZIhvcNAQEFBQAwVzELMAkGA1UEBhMCQkUx
GTAXBgNVBAoTEEdsb2JhbFNpZ24gbnYtc2ExEDAOBgNVBAsTB1Jvb3QgQ0ExGzAZBgNVBAMTEkds
b2JhbFNpZ24gUm9vdCBDQTAeFw05ODA5MDExMjAwMDBaFw0yODAxMjgxMjAwMDBaMFcxCzAJBgNV
BAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWduIG52LXNhMRAwDgYDVQQLEwdSb290IENBMRswGQYD
VQQDExJHbG9iYWxTaWduIFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDa
DuaZjc6j40+Kfvvxi4Mla+pIH/EqsLmVEQS98GPR4mdmzxzdzxtIK+6NiY6arymAZavpxy0Sy6sc
THAHoT0KMM0VjU/43dSMUBUc71DuxC73/OlS8pF94G3VNTCOXkNz8kHp1Wrjsok6Vjk4bwY8iGlb
Kk3Fp1S4bInMm/k8yuX9ifUSPJJ4ltbcdG6TRGHRjcdGsnUOhugZitVtbNV4FpWi6cgKOOvyJBNP
c1STE4U6G7weNLWLBYy5d4ux2x8gkasJU26Qzns3dLlwR5EiUWMWea6xrkEmCMgZK9FGqkjWZCrX
gzT/LCrBbBlDSgeF59N89iFo7+ryUp9/k5DPAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNV
HRMBAf8EBTADAQH/MB0GA1UdDgQWBBRge2YaRQ2XyolQL30EzTSo//z9SzANBgkqhkiG9w0BAQUF
AAOCAQEA1nPnfE920I2/7LqivjTFKDK1fPxsnCwrvQmeU79rXqoRSLblCKOzyj1hTdNGCbM+w6Dj
Y1Ub8rrvrTnhQ7k4o+YviiY776BQVvnGCv04zcQLcFGUl5gE38NflNUVyRRBnMRddWQVDf9VMOyG
j/8N7yy5Y0b2qvzfvGn9LhJIZJrglfCm7ymPAbEVtQwdpf5pLGkkeB6zpxxxYu7KyJesF12KwvhH
hm4qxFYxldBniYUr+WymXUadDKqC5JlR3XC321Y9YeRq4VzW9v493kHMB65jUr9TU/Qr6cf9tveC
X4XSQRjbgbMEHMUfpIBvFSDJ3gyICh3WZlXi/EjJKSZp4A==
-----END CERTIFICATE-----`,
			},
		),
	}

	cmd.Flags().String("description", "", "Description of the certificate authority.")
	cmd.Flags().String("certificate-chain", "", "A base64 encoded string containing the signing certificate chain.")
	cmd.Flags().String("certificate-chain-filename", "", "The name of the certificate file.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("description"))
	cobra.CheckErr(cmd.MarkFlagRequired("certificate-chain"))
	cobra.CheckErr(cmd.MarkFlagRequired("certificate-chain-filename"))

	return cmd
}

func (c *certificateAuthorityCommand) create(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	certificateChain, err := cmd.Flags().GetString("certificate-chain")
	if err != nil {
		return err
	}

	certificateChainFilename, err := cmd.Flags().GetString("certificate-chain-filename")
	if err != nil {
		return err
	}

	certRequest := certificateauthorityv2.IamV2CreateCertRequest{
		DisplayName:              certificateauthorityv2.PtrString(args[0]),
		Description:              certificateauthorityv2.PtrString(description),
		CertificateChain:         certificateauthorityv2.PtrString(certificateChain),
		CertificateChainFilename: certificateauthorityv2.PtrString(certificateChainFilename),
	}

	certificateAuthority, err := c.V2Client.CreateCertificateAuthority(certRequest)
	if err != nil {
		return err
	}

	return printCertificateAuthority(cmd, certificateAuthority)
}
