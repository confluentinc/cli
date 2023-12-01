package byok

import (
	"fmt"
	"regexp"
	"strings"
)

type gcpPolicyMetadata struct {
	project  string
	location string
	keyRing  string
	key      string
	group    string
}

func (g gcpPolicyMetadata) renderPolicy() string {
	return fmt.Sprintf(`
gcloud iam roles create %[1]s \
	--project=%[2]s \
	--description="Grant necessary permissions for Confluent to access KMS key" \
	--permissions=cloudkms.cryptoKeyVersions.useToDecrypt,cloudkms.cryptoKeyVersions.useToEncrypt,cloudkms.cryptoKeys.get && \
gcloud kms keys add-iam-policy-binding %[3]s \
	--project=%[2]s \
	--keyring="%[4]s" \
	--location="%[5]s" \
	--member="group:%[6]s" \
	--role="projects/%[2]s/roles/%[1]s"`, g.getCustomRoleName(), g.project, g.key, g.keyRing, g.location, g.group)
}

func (g gcpPolicyMetadata) getCustomRoleName() string {
	r := regexp.MustCompile(`^[a-zA-Z0-9_.]{3,64}$`)

	customRoleName := strings.ReplaceAll(fmt.Sprintf("%s_%s_custom_kms_role", g.keyRing, g.key), "-", "_")
	if ok := r.Match([]byte(customRoleName)); !ok {
		return defaultGcpRoleName
	}

	return customRoleName
}
