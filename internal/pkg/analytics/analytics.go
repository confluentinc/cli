package analytics

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	segment "github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	secretCommandFlags    = map[string][]string{
		"ccloud init": {"api-secret"},
		"confluent master-key generate": {"passphrase", "local-secrets-file"},
		"confluent file rotate": {"passphrase", "passphrase-new"},
	}
	secretCommandArgs     = map[string][]int{"ccloud api-key store": {1}}
	SecretValueString     = "<secret_value>"
	malformedCmdEventName = "Malformed Command Error"

	// these are exported to avoid import cycle with test (test is in package analytics_test)
	FlagsPropertiesKey      = "flags"
	ArgsPropertiesKey       = "args"
	OrgIdPropertiesKey      = "organization_id"
	EmailPropertiesKey      = "email"
	ErrorMsgPropertiesKey   = "error_message"
	StartTimePropertiesKey  = "start_time"
	FinishTimePropertiesKey = "finish_time"
	SucceededPropertiesKey  = "succeeded"
	CredentialPropertiesKey = "credential"
	ApiKeyPropertiesKey     = "apikey"
	VersionPropertiesKey    = "version"
	CliNameTraitsKey        = "cli_name"

	apiKeyCred   = "apikey"
	userNameCred = "username"

	tokenExpiredErrorMessage = errors.TypeMessages[reflect.TypeOf(&ccloud.ExpiredTokenError{})]
)

type Client struct {
	client      segment.Client
	config      *config.Config

	// cache data until we flush events to segment (when each cmd call finishes)
	cmdCalled   string
	properties  segment.Properties
	userId      string
	anonymousId string
	cliVersion  string
}

func NewAnalyticsClient(cfg *config.Config, version string, segmentClient segment.Client) *Client {
	client := &Client{
		client:     segmentClient,
		config:     cfg,
		properties: make(segment.Properties),
		cliVersion:    version,
	}
	if cfg.AnonymousId == "" {
		cfg.AnonymousId = uuid.New().String()
	}
	client.anonymousId = cfg.AnonymousId
	return client
}

func (a *Client) TrackCommand(cmd *cobra.Command, args []string) {
	a.cmdCalled = cmd.CommandPath()
	a.addArgsProperties(cmd, args)
	a.addFlagProperties(cmd)
	a.properties.Set(StartTimePropertiesKey, time.Now())
	a.properties.Set(VersionPropertiesKey, a.cliVersion)
	a.setUserInfo()
}

func (a *Client) FlushCommandSucceeded() error {
	// for login user info can only be obtained after login succeeded
	if strings.Contains(a.cmdCalled, a.config.CLIName + " login") {
		a.setUserInfo()
		if err := a.identify(); err != nil {
			return err
		}
	}
	a.properties.Set(SucceededPropertiesKey, true)
	a.properties.Set(FinishTimePropertiesKey, time.Now())
	if err := a.sendPage(); err != nil {
		return err
	}
	if strings.Contains(a.cmdCalled, a.config.CLIName + " logout") {
		if err := a.resetAnonymousId(); err != nil {
			return err
		}
	}
	return a.client.Close()
}

func (a *Client) FlushCommandFailed(e error) error {
	if a.cmdCalled == "" {
		return a.malformedCommandError(e)
	}
	if e.Error() == tokenExpiredErrorMessage{
		if err := a.resetAnonymousId(); err != nil {
			return err
		}
	}
	a.properties.Set(SucceededPropertiesKey, false)
	a.properties.Set(ErrorMsgPropertiesKey, e.Error())
	if err := a.sendPage(); err != nil {
		return err
	}
	return a.client.Close()
}

func (a *Client) sendPage() error {
	page := segment.Page{
		AnonymousId:  a.anonymousId,
		Name:         a.cmdCalled,
		Properties:   a.properties,
	}
	if a.userId != "" {
		page.UserId = a.userId
	}
	cred := a.getCredentialType()
	a.properties.Set(CredentialPropertiesKey, cred)
	if cred == apiKeyCred {
		a.properties.Set(ApiKeyPropertiesKey, a.getCredApiKey())
	}
	return a.client.Enqueue(page)
}

func (a *Client) identify() error {
	identify := segment.Identify{
		AnonymousId:  a.anonymousId,
		UserId:       a.userId,
	}
	traits := segment.Traits{}
	traits.Set(VersionPropertiesKey, a.cliVersion)
	traits.Set(CliNameTraitsKey, a.config.CLIName)
	identify.Traits = traits
	return a.client.Enqueue(identify)
}

func (a *Client) malformedCommandError(e error) error {
	properties := make(segment.Properties)
	properties.Set(ErrorMsgPropertiesKey, e.Error())
	track := segment.Track{
		AnonymousId: a.anonymousId,
		Event:       malformedCmdEventName,
		Properties:  properties,
	}
	if a.userId != "" {
		track.UserId = a.userId
	}
	err := a.client.Enqueue(track)
	if err != nil {
		return err
	}
	return a.client.Close()
}

func (a *Client) addFlagProperties(cmd *cobra.Command) {
	flags := make(map[string]string)
	cmd.Flags().Visit(func(f *pflag.Flag){
		if flagNames, ok := secretCommandFlags[cmd.CommandPath()]; ok {
			for _, flagName := range flagNames {
				if f.Name == flagName {
					flags[f.Name] = SecretValueString
					break
				}
			}
		}
		if _, ok := flags[f.Name]; !ok {
			flags[f.Name] = f.Value.String()
		}
	})
	a.properties.Set(FlagsPropertiesKey, flags)
}

func  (a *Client) addArgsProperties(cmd *cobra.Command, args []string) {
	argsCopy := make([]string, len(args))
	if len(args) > 0 {
		if ids, ok := secretCommandArgs[cmd.CommandPath()]; ok {
			copy(argsCopy, args)
			for _, i := range ids {
				argsCopy[i] = SecretValueString
			}
		} else {
			copy(argsCopy, args)
		}
	}
	a.properties.Set(ArgsPropertiesKey, argsCopy)
}

func (a *Client) setUserInfo() {
	if a.config.CLIName == "ccloud" {
		cloudUserId, organizationId, email := a.getCloudUserInfo()
		if cloudUserId != "" {
			a.properties.Set(OrgIdPropertiesKey, organizationId)
			a.properties.Set(EmailPropertiesKey, email)
		}
		a.userId = cloudUserId
	} else {
		a.userId = a.getCPUsername()
	}
}

func (a *Client) getCloudUserInfo() (userId string, organizationId string, email string) {
	if err := a.config.CheckLogin(); err != nil {
		return "", "", ""
	}
	fmt.Println("HA IM LOGGED IN")
	user := a.config.Auth.User
	userId = strconv.Itoa(int(user.Id))
	organizationId = strconv.Itoa(int(user.OrganizationId))
	email = user.Email
	return userId, organizationId, email
}

func (a *Client) getCPUsername() string {
	if err := a.config.CheckLogin(); err != nil {
		return ""
	}
	ctx := a.config.Contexts[a.config.CurrentContext]
	cred := a.config.Credentials[ctx.Credential]
	return cred.Username
}

func (a *Client) getCredentialType() string {
	credType, err := a.config.CredentialType()
	if err != nil {
		return "none"
	}
	switch credType {
	case config.Username:
		if a.config.CheckLogin() == nil {
			return userNameCred
		}
	case config.APIKey:
		return apiKeyCred
	}
	return "none"
}

func (a *Client) getCredApiKey() string {
	context, err := a.config.Context()
	if err != nil {
		return ""
	}
	if cred, ok := a.config.Credentials[context.Credential]; ok {
		return cred.APIKeyPair.Key
	}
	return ""
}

func (a *Client) resetAnonymousId() error {
	a.config.AnonymousId = ""
	if err := a.config.Save(); err != nil {
		return err
	}
	return nil
}
