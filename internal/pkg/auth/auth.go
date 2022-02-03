package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/dghubble/sling"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	CCloudURL = "https://confluent.cloud"

	ConfluentCloudEmail          = "CONFLUENT_CLOUD_EMAIL"
	ConfluentCloudPassword       = "CONFLUENT_CLOUD_PASSWORD"
	ConfluentCloudOrganizationId = "CONFLUENT_CLOUD_ORGANIZATION_ID"
	ConfluentPlatformUsername    = "CONFLUENT_PLATFORM_USERNAME"
	ConfluentPlatformPassword    = "CONFLUENT_PLATFORM_PASSWORD"
	ConfluentPlatformMDSURL      = "CONFLUENT_PLATFORM_MDS_URL"
	ConfluentPlatformCACertPath  = "CONFLUENT_PLATFORM_CA_CERT_PATH"

	DeprecatedConfluentCloudEmail         = "CCLOUD_EMAIL"
	DeprecatedConfluentCloudPassword      = "CCLOUD_PASSWORD"
	DeprecatedConfluentPlatformUsername   = "CONFLUENT_USERNAME"
	DeprecatedConfluentPlatformPassword   = "CONFLUENT_PASSWORD"
	DeprecatedConfluentPlatformMDSURL     = "CONFLUENT_MDS_URL"
	DeprecatedConfluentPlatformCACertPath = "CONFLUENT_CA_CERT_PATH"
)

// GetEnvWithFallback calls os.GetEnv() twice, once for the current var and once for the deprecated var.
func GetEnvWithFallback(current, deprecated string) string {
	if val := os.Getenv(current); val != "" {
		return val
	}

	if val := os.Getenv(deprecated); val != "" {
		_, _ = fmt.Fprintf(os.Stderr, errors.DeprecatedEnvVarWarningMsg, deprecated, current)
		return val
	}

	return ""
}

func PersistLogoutToConfig(config *v1.Config) error {
	ctx := config.Context()
	if ctx == nil {
		return nil
	}
	err := ctx.DeleteUserAuth()
	if err != nil {
		return err
	}
	ctx.Config.CurrentContext = ""
	return config.Save()
}

func PersistConfluentLoginToConfig(config *v1.Config, username string, url string, token string, caCertPath string, isLegacyContext bool) error {
	state := &v1.ContextState{
		Auth:      nil,
		AuthToken: token,
	}
	var ctxName string
	if isLegacyContext {
		ctxName = GenerateContextName(username, url, "")
	} else {
		ctxName = GenerateContextName(username, url, caCertPath)
	}
	return addOrUpdateContext(config, ctxName, username, url, state, caCertPath, "")
}

func PersistCCloudLoginToConfig(config *v1.Config, email string, url string, token string, client *ccloud.Client) (*orgv1.Account, error) {
	ctxName := GenerateCloudContextName(email, url)
	user, err := getCCloudUser(client)
	if err != nil {
		return nil, err
	}
	state := getCCloudContextState(config, ctxName, token, user)

	err = addOrUpdateContext(config, ctxName, email, url, state, "", user.Organization.ResourceId)
	return state.Auth.Account, err
}

func addOrUpdateContext(config *v1.Config, ctxName string, username string, url string, state *v1.ContextState, caCertPath, orgResourceId string) error {
	credName := generateCredentialName(username)
	platform := &v1.Platform{
		Name:       strings.TrimPrefix(url, "https://"),
		Server:     url,
		CaCertPath: caCertPath,
	}
	credential := &v1.Credential{
		Name:     credName,
		Username: username,
		// don't save password if they entered it interactively.
	}

	if err := config.SavePlatform(platform); err != nil {
		return err
	}

	if err := config.SaveCredential(credential); err != nil {
		return err
	}

	if ctx, ok := config.Contexts[ctxName]; ok {
		config.ContextStates[ctxName] = state
		ctx.State = state

		ctx.Platform = platform
		ctx.PlatformName = platform.Name

		ctx.Credential = credential
		ctx.CredentialName = credential.Name
		ctx.LastOrgId = orgResourceId
	} else {
		if err := config.AddContext(ctxName, platform.Name, credential.Name, map[string]*v1.KafkaClusterConfig{}, "", nil, state, orgResourceId); err != nil {
			return err
		}
	}

	return config.UseContext(ctxName)
}

func getCCloudContextState(config *v1.Config, ctxName, token string, user *flowv1.GetMeReply) *v1.ContextState {
	var state *v1.ContextState
	ctx, err := config.FindContext(ctxName)
	if err == nil {
		state = ctx.State
	} else {
		state = new(v1.ContextState)
	}
	state.AuthToken = token

	if state.Auth == nil {
		state.Auth = &v1.AuthConfig{}
	}

	// Always overwrite the user, organization, and list of accounts when logging in -- but don't necessarily
	// overwrite `Account` (current/active environment) since we want that to be remembered
	// between CLI sessions.
	state.Auth.User = user.User
	state.Auth.Accounts = user.Accounts
	state.Auth.Organization = user.Organization

	// Default to 0th environment if no suitable environment is already configured
	hasGoodEnv := false
	if state.Auth.Account != nil {
		for _, acc := range state.Auth.Accounts {
			if acc.Id == state.Auth.Account.Id {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		state.Auth.Account = state.Auth.Accounts[0]
	}

	return state
}

func getCCloudUser(client *ccloud.Client) (*flowv1.GetMeReply, error) {
	user, err := client.Auth.User(context.Background())
	if err != nil {
		return nil, err
	}
	if len(user.Accounts) == 0 {
		return nil, errors.Errorf(errors.NoEnvironmentFoundErrorMsg)
	}
	return user, nil
}

func GenerateCloudContextName(username string, url string) string {
	return GenerateContextName(username, url, "")
}

// if CP users use cacertpath then include that in the context name
// (legacy CP users may still have context without cacertpath in the name but have cacertpath stored)
func GenerateContextName(username string, url string, caCertPath string) string {
	if caCertPath == "" {
		return fmt.Sprintf("login-%s-%s", username, url)
	}
	return fmt.Sprintf("login-%s-%s?cacertpath=%s", username, url, caCertPath)
}

func generateCredentialName(username string) string {
	return fmt.Sprintf("username-%s", username)
}

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func GetBearerToken(authenticatedState *v1.ContextState, server, clusterId string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"

	s := map[string][]string{"clusterIds": {clusterId}}
	fmt.Println(s)

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	_, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).BodyJSON(s).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}
	//return "eyJhbGciOiJSUzI1NiIsImprdSI6Imh0dHBzOi8vYXV0aC1zdGF0aWMuY29uZmx1ZW50LmlvL2p3a3MiLCJraWQiOiIxMjc5ZDg4My1kNjBhLTExZWItOTZlMC0xMjZiOTk2MmM4YjAiLCJ0eXAiOiJKV1QifQ.eyJjbHVzdGVycyI6WyJsY2MtNnI3bjIiLCJsY2MtemR5ZDMiLCJsa2MtOHhncjAiLCJsY2MtcTA1cG0iLCJsY2MtM3gzcTIiLCJsY2MtMXBkejYiLCJsY2MtNnI3cWoiLCJsY2MtZG4weDciLCJsY2MtM3gzNWoiLCJsY2MtcDI1NXkiLCJsY2Mtbmo1amsiLCJsa2MtMXgwbjMiLCJsY2MteGR5ZHoiLCJsY2Mtd2R5MG0iLCJsY2MtODBxcHIiLCJsY2MtMXBkZHYiLCJsY2MtNzJvenciLCJsY2MtanB3bjIiLCJsY2MtNWoxem4iLCJsY2MtbTk1d3ciLCJsY2Mtbmo1bXYiLCJsY2MteWQ5M2siLCJsY2Mtcjh6bTEiLCJsY2MtdmR5NXAiLCJsY2MteGR5NXEiLCJsY2Mta3Y2d20iLCJsY2Mtb3I1NjkiLCJsY2MtZG4wOHkiLCJsY2MtemR5bzMiLCJsY2MtcTA1NjIiLCJsY2MtZ294ZDMiLCJsY2MtZ294NW4iLCJsY2MtOThnMnkiLCJsY2Mtcjh6ODkiLCJsY2Mtb3I1bXkiLCJsa2MtM3p3bTIiLCJsY2MtcTA1MDYiLCJsY2MtM3gzMGoiLCJsY2MtMXBkMHYiLCJsY2Mtbmo1MGQiLCJsY2MteGR5MWciLCJsY2MtanB3bTIiLCJsY2MtcTA1NW0iLCJsY2MteGR5a3giLCJsY2MtZDZ3MGQiLCJsY2MtZG4wbm8iLCJsY2MtZ294bzMiLCJsY2Mtb3I1d3giLCJsY2Mtd2R5ZDkiLCJsY2MtMXBkMHoiLCJsY2MtMGc4ejIiLCJsY2MtMGc4Z3AiLCJsY2MtNnI3eGoiLCJsY2MtODBxcnIiLCJsY2MtNWoxZzgiLCJsY2MtNWtkajgiLCJsY2MteGR5bmsiLCJsY2Mtb3I1djkiLCJsY2MtNWoxeGciLCJsY2Mtcjh6NmsiLCJsY2MteGR5MHgiLCJsY2MtMGc4ODUiLCJsY2MtNnI3OHEiLCJsY2MtdmR5ZHoiLCJsY2MtMXBka3oiLCJsY2MtMXBkODYiLCJsY2MtanB3MnciLCJsY2MtMXBkODUiLCJsY2MtcTA1bTciLCJsY2MtOThnZzciLCJsY2MtMno4em8iLCJsY2MtMno4NnkiLCJsY2MteWQ5NWoiLCJsY2MtODBxNjciLCJsY2Mtcjh6Z2siLCJsY2MtbTk1NngiLCJsY2Mta3Y2bXYiLCJsY2MteWQ5a2oiLCJsY2Mta3Y2NWciLCJsY2Mtd2R5a20iLCJsY2MteWQ5ZDciLCJsY2MtNzJvdm8iLCJsY2Mtbmo1emsiLCJsY2MteWQ5MXAiLCJsY2MtNnI3OWoiLCJsY2MtNWoxeTIiLCJsY2Mtem9qMnkiLCJsY2MtNnI3MGoiLCJsY2MtanB3eHAiLCJsY2MtODBxa3EiLCJsY2MtODBxMDAiLCJsY2MtcDI1cW8iLCJsY2MtbTk1cXEiLCJsY2MtNnI3enEiLCJsY2Mta3Y2M3AiLCJsY2MtMXBkeGoiLCJsY2MtanB3MTIiLCJsY2MtZG4wM2QiLCJsY2MtMGc4azYiLCJsY2MtNzJveHAiLCJsY2MtNzJvZHciLCJsY2Mtbmo1ZHYiLCJsY2MtdmR5MDUiLCJsY2MtM3gzNm8iLCJsY2MtMGc4MHEiLCJsY2MtNzJvM28iLCJsY2Mta3Y2N3YiLCJsY2MtbTlxanEiLCJsY2Mtbmo1cHoiLCJsY2MtMGc4bnAiLCJsY2MtODBxbjUiLCJsY2MtNzJvbzIiLCJsY2Mtcjh6cDkiLCJsY2MteWRucnAiLCJsY2MteGR5MnoiLCJsY2MtODBxeDAiLCJsY2MtcTA1cTIiLCJsY2MtODBxNzciLCJsY2MtcTA1ZzYiLCJsY2MtcTA1bjIiLCJsY2MtZ294Z3IiLCJsY2MtNnI3NTMiLCJsY2Mtb3I1ajkiLCJsY2MtcDI1ZG8iLCJsY2MtZ294N24iLCJsY2MtdmR5b3oiLCJsY2MtcDI1M2siLCJsY2Mtcjh6MzkiLCJsY2MtMGc4MXAiLCJsY2MtZG4wMjEiLCJsY2MtanB3Z3ciLCJsY2MtemR5MnkiLCJsY2MtM3gzdjAiLCJsY2MteWQ5eGoiLCJsY2Mtd2R5cG0iLCJsY2MtNWoxbjIiLCJsY2MtMXBkMTUiLCJsY2MteGR5Z3oiLCJsY2Mtbmo1NWQiLCJsY2MtNnI3MjMiLCJsY2Mta3Y2b20iLCJsY2MtOThneTUiLCJsY2Mtbmo1OXoiLCJsY2Mtbmo1MXoiLCJsY2MtcDI1em8iLCJsY2MtdmR5cXoiLCJsY2Mtb3I1bngiLCJsY2MtZ294eTEiLCJsY2MtNWc2MTIiLCJsY2MtdmR5djUiLCJsY2MtODBxODAiLCJsY2MteWQ5djciLCJsY2MteGR5emsiLCJsY2Mtbmo1a2QiLCJsY2MtMXBkbnYiLCJsY2MtcDI1NmsiLCJsY2MtanB3eTIiLCJsY2MtMno4cXEiLCJsY2Mta3Y2cnAiLCJsY2MtODBxanIiLCJsY2MtcDI1ankiLCJsY2Mtd2R5eGciLCJsY2MtMno4NTEiLCJsY2MtODBxeXEiLCJsY2MtcDI1Mm8iLCJsY2MtemR5cjAiLCJsY2Mtbmo1cXYiLCJsY2Mtb3I1Z3giLCJsY2MtcTA1dnAiLCJsY2MtNWoxNm4iLCJsY2MtZ294cjEiLCJsY2MtZG4wd3kiLCJsY2MtemR5ajciLCJsY2MtMXBkNXYiLCJsY2MtZ294MW0iLCJsY2MtanB3dzgiLCJsY2Mtcjh6ejAiLCJsY2MtMno4ODEiLCJsY2Mtd2R5eXciLCJsY2MtODBxcXEiLCJsY2MtbTk1NXciLCJsY2Mta3Y2NjIiLCJsY2MtZG4wMHkiLCJsY2MtNnI3N3EiLCJsY2MtemR5eTciLCJsc3JjLW9nMThwIiwibGNjLXhkeXd4IiwibGNjLXI4endrIiwibGNjLW5qNTdrIiwibGNjLWRuMHBkIiwibGtjLW9nMHBqIiwibGNjLTN4M21vIiwibGNjLTBnOG02IiwibGNjLXI4ejFrIiwibGNjLTgwcW83IiwibGNjLXlkOTA3IiwibGNjLW9yNTBqIiwibGNjLXpkeTY3IiwibGNjLXZkeThwIiwibGNjLTVqMTBxIiwibGNjLWpwd2RtIiwibGNjLXI4engxIiwibGNjLTN4M3p3IiwibHNyYy1xdm9rZCIsImxjYy1kbjB6ZCIsImxjYy1tOTVvcSIsImxjYy1tOTUxMSIsImxjYy0xcGR2NSIsImxjYy1kbjB5NyIsImxjYy05OGcxbSIsImxjYy1rdjYxdiIsImxjYy05OGc4diIsImxjYy03Mm84cCIsImxjYy1kbjA2byIsImxjYy05OGdkNSIsImxjYy0zeDNwMiIsImxjYy1xMDV4cCIsImxjYy01ajFyOCIsImxjYy0yejhncSIsImxjYy03Mm8ybyIsImxjYy0zeDNqMiIsImxjYy1rdjYwMiIsImxjYy1vcjVyeSIsImxjYy15ZDk5ayIsImxjYy03Mm9ncCIsImxjYy03Mm8xdyIsImxjYy13ZHl3bSIsImxjYy15ZDlwNyIsImxjYy0xcGRxaiIsImxjYy0zeDN5MCIsImxjYy0xcGQ3aiIsImxjYy1uajV6NiIsImxjYy0yejg3byIsImxjYy0zeDNxdyIsImxjYy1qcHd6OCIsImxjYy1vcjVveSIsImxjYy03Mm9qcCIsImxjYy1nb3gwMSIsImxjYy0yejl2cSIsImxjYy1tOTV4eCIsImxjYy05OGdxNSIsImxjYy1nejFqbiIsImxjYy1wMjVtNSIsImxjYy1tOTVneCIsImxjYy1kbjBtMSIsImxjYy03Mm8zMSIsImxjYy02cjdqOCIsImxjYy03Mm81MiIsImxjYy0wZzg2NiIsImxjYy0yNzI4cSIsImxjYy1qcHc2bSIsImxjYy1tOTV5cSIsImxjYy0xMTNkMyIsImxjYy02cDg3MiIsImxjYy0xMTVkeiIsImxjYy0zbzUzbyIsImxjYy1rM3g2bSIsImxjYy14ZHl2eCIsImxjYy01ajExbiIsImxjYy03Mm85cCIsImxjYy1nb3h4bSIsImxjYy1uajVuNiIsImxjYy13ZHlqdyIsImxjYy15ZDk4cCIsImxjYy1uajVndiIsImxjYy13ZHkzZyIsImxjYy16ZHl4MyIsImxjYy1uajU4ayIsImxjYy1xMDU3cCIsImxjYy05OGdqdiIsImxjYy1tOTVtcSIsImxjYy1rdjZ2cCIsImxzcmMtOWd3a3YiLCJsa2MtbnFqM2siLCJsc3JjLW9nOTBqIiwibGNjLXB5cnc1IiwibGNjLW43cnJ6IiwibGNjLTJ6OG9vIiwibGNjLXFnM2o2IiwibGNjLXpkeTEwIiwibGNjLXI4ejkxIiwibGNjLTd2Z29qIiwibGNjLXEwNWQ2IiwibGNjLXdkeW1nIiwibGNjLXpkeXdkIiwibGNjLXhkeTNrIiwibGNjLXlkOTJrIiwibGNjLXI4enlwIiwibGNjLTgwMzVyIiwibGNjLTF6bzE1IiwibGNjLTdyOXZvIiwibGNjLXI4em8xIiwibGNjLWRuMHFvIiwibGNjLXZkeXlwIiwibGNjLTk4Z281IiwibGNjLTBnOHYyIiwibGNjLTk4Z255IiwibGNjLXlkOW5wIiwibGNjLTVqMW9nIiwibGNjLXAyNXA1IiwibGNjLXhkeXB6IiwibGNjLXZkeXduIiwibGNjLTFwZDZ6IiwibGNjLTBnOHBwIiwibGNjLTN4M3l3IiwibGNjLTFwZG16IiwibGNjLTcybzd3IiwibGNjLTVqMXZnIiwibGNjLTBnOHk2IiwibGNjLTJ6OG1tIiwibGNjLTN4M2RvIiwibGNjLTN4M293IiwibGNjLXhkeThrIiwibGNjLW05NTkxIiwibGNjLXdkeWc5IiwibGNjLXI4em45IiwibGNjLTBnODU1IiwibGNjLXAyNW9rIiwibGNjLTJ6OHhvIiwibGNjLXlkOXdqIiwibGNjLTJ6OGR5IiwibGNjLW9yNWR5IiwibGNjLW05NXB3IiwibGNjLXlkOXpvIiwibGNjLTZyNzEyIiwibGNjLTFwZHl6IiwibGNjLWpwd3BtIiwibGNjLXZkeTluIiwibGNjLXpkeTlkIiwibGNjLWpwdzd3IiwibGNjLXAyNTE1IiwibGNjLXdkeXI5IiwibGNjLW9yNXB4IiwibGNjLTN4M3h3IiwibGNjLTVqMWs4IiwibGNjLTFwZG9qIiwibGNjLTJ6ODlxIiwibGNjLTk4Z3Z5IiwibGNjLTk4Z3J5IiwibGNjLWt2NjJtIiwibGNjLWt2Nnl2IiwibGNjLXpkeTcwIiwibGNjLW05NTM3IiwibGNjLXZkeTMwIiwibGNjLWdveHozIiwibGNjLWRuMHI3IiwibGNjLW5qNTZ6IiwibGNjLXAyNXhrIiwibGNjLXdkeTZ3IiwibGNjLTFwZHA1IiwibGNjLTcyb3oyIiwibGNjLTVqMTMyIiwibGNjLTBnODcyIiwibGNjLTk4ZzM3IiwibGNjLTcyb2syIiwibGNjLW9yNTVvIiwibGNjLWRuMGo3IiwibGNjLWdveDltIiwibGNjLWpwdzA4IiwibGNjLTJ6OG5xIiwibGNjLTN4M2tvIiwibGNjLWdveDJuIiwibGNjLW05NTAxIiwibGNjLTcyb213IiwibGNjLWdveDNuIiwibGNjLXZkeWp6IiwibGNjLXpkeWtkIiwibGNjLXdkeXE5IiwibGNjLW5qNWRrIiwibGNjLTk4Z212IiwibGNjLWt2NnptIiwibGNjLW5qNTB6IiwibGNjLTN4M2cyIiwibGNjLWpwd3J3IiwibGNjLXZkeWtuIiwibGNjLXpkeW43IiwibGNjLXdkeThnIiwibGNjLW9yNXpvIiwibGNjLXdkeTk1IiwibGNjLW5qNXJkIiwibGNjLW05NWQxIiwibGNjLXEwNTM2IiwibGNjLWpwd2ptIiwibGNjLTgwcTNyIiwibGNjLTFwZDNqIiwibGNjLTJ6OGt5IiwibGNjLWdveDgxIiwibGNjLTZyNzYyIiwibGNjLXpkeTUwIiwibGNjLXAyNThtIiwibGNjLXpkeTgzIiwibGNjLXAyNXY1IiwibGNjLXI4enYwIiwibGNjLXI4emswIiwibGNjLTgwcTkwIiwibGNjLW5qNTN2IiwibGNjLXEwNXlwIiwibGNjLTVqMWQyIiwibGNjLXAyNXd5IiwibGNjLXEwNXcyIiwibGNjLXhkeXlxIiwibGNjLW9yNTc5IiwibGNjLXZkeTI1IiwibGNjLTVqMTJnIiwibGNjLTZyN3AzIiwibGNjLTVqMTlxIiwibGNjLTcyb3hvIiwibGNjLTZyN2dxIiwibGNjLTN4MzEyIiwibGNjLXEwNXptIiwibGNjLTZyN3kyIiwibGNjLTk4Zzl2IiwibGNjLTgwcWQ3IiwibGNjLW05NW54IiwibGNjLTk4ZzU3IiwibGNjLWRuMDVkIiwibGNjLTZyN3IzIiwibGNjLTVqMWo4IiwibGNjLTN4MzNqIiwibGNjLTJ6ODAxIiwibGNjLTJ6OHl5IiwibGNjLTBnODIyIiwibGNjLTN4MzBvIiwibGNjLTN4MzdqIiwibGNjLTBnOHc1IiwibGNjLTN2cW1qIiwibGNjLTd2a293IiwibGNjLW43cjV6IiwibGtjLXZ3cDcwIl0sIm9yZ2FuaXphdGlvbklkIjozNTc2MCwidXNlcklkIjoyMjA0NjksInVzZXJSZXNvdXJjZUlkIjoidS1lMDJ4MjIiLCJleHAiOjE2NDMyMzMyNDUsImp0aSI6Ijc1MGZiODEwLTZlMzktNDU2YS05YzVmLTFhOGJjZGU5ZmNjOCIsImlhdCI6MTY0MzIzMjM0NSwiaXNzIjoiQ29uZmx1ZW50Iiwic3ViIjoiMjIwNDY5In0.Q9hCW43Kwn92zDRjTpLqqEsEcT8CQS33IwTU_hGJ-j1VGA3su3Br5jByG07Pe2MFiEW5yEmZQezo5oRR22I3y_66Vc9R1_fuudEcUMODvJxyST5kL0IZe40dYhh6li3bHrbMFtbFvP0c2fv9P6MHj5HQzZeZBRiJkHYjvQR_xOtGD73dJSzNNj51jMaDUOJRydXNnZpIPMDJ157ixlY3PB3aACA1oMHze8mG4CjjMJgLd3MlVyJgyaSM55W4p0A2ex5wXBslCVKw2hHCfcPBhfLTyCU1IyDPZqEulWyjkQDFvI44zRCE9UzVxvxoqd3aQD8z_IywhC9yH8eVwBtI8w", nil
	fmt.Println("TOKEN IS")
	fmt.Println(responses.Token)
	return responses.Token, nil
}
