package testserver

import (
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"
)

const initialKeyIndex = int32(3)

// ResetState restores the data that handlers mutate at runtime to its seed values. Any new
// package-level variable that a handler writes to must be reset here, or tests become
// order-dependent.
func ResetState() {
	keyIndex = initialKeyIndex
	keyStoreV2 = map[string]*apikeysv2.IamV2ApiKey{}
	fillKeyStoreV2()
	artifactStore = map[string]camv1.CamV1ConnectArtifact{}
}
