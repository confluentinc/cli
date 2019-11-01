package analytics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joeshaw/iso8601"
	"github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/config"
)

var (
	secretFlags = []string{"flag1", "flag2"}
	secretCommands = map[string][]int{"ccloud api-key store": {1}} // security concern IF there is a change then we could accidentally be sending secret value to segment
	secretValueString = "<secret_value>"
)

type Object struct {
	client     analytics.Client
	config     *config.Config
	cmdCalled  string
	properties analytics.Properties
	userId     string
}

func NewAnalyticsObject(cfg *config.Config) *Object {
	obj := &Object{
		client: analytics.New("waLqtpvaj5o0YKOQGi7gOgav9gIi9oCJ"),
		config: cfg,
		properties: make(analytics.Properties),
	}
	if cfg.AnonymousId == "" {
		cfg.AnonymousId = uuid.New().String()
	}
	return obj
}

func (a *Object) InitializePreRuns(cmd *cobra.Command) error {
	if !cmd.HasSubCommands() {
		if cmd.PreRun == nil {
			cmd.PreRun = a.PreRun()
		} else {
			return fmt.Errorf("pre run already existed") // preventing preRun collisions
		}
		return nil
	}
	for _, childCmd := range cmd.Commands() {
		err := a.InitializePreRuns(childCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Object) PreRun() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		a.cmdCalled = cmd.CommandPath()
		a.addArgsProperties(cmd, args)
		a.addFlagProperties(cmd)
		a.properties.Set("start_time", iso8601.New(time.Now()))
	}
}

func (a *Object) FlushCommandSucceeded() error {
	a.setUserInfo()
	a.properties.Set("succeeded", true)
	a.properties.Set("finish_time", iso8601.New(time.Now()))
	if err := a.sendPage(); err != nil {
		return err
	}
	// reset anonymous id
	if strings.Contains(a.cmdCalled, a.config.CLIName + " logout") {
		a.config.AnonymousId = ""
	}
	// identify the user with the anonymous id if login called
	if a.userId != "" && strings.Contains(a.cmdCalled, a.config.CLIName + " login") {
		if err := a.identify(); err != nil {
			return err
		}
	}
	return a.client.Close()
}

func (a *Object) FlushCommandFailed(e error) error {
	a.setUserInfo()
	if a.cmdCalled == "" {
		return a.malformedCommandError(e)
	}
	a.properties.Set("succeeded", false)
	a.properties.Set("error_message", e.Error())
	if err := a.sendPage(); err != nil {
		return err
	}
	return a.client.Close()
}

func (a *Object) sendPage() error {
	page := analytics.Page{
		AnonymousId:  a.config.AnonymousId,
		Name:         a.cmdCalled,
		Properties:   a.properties,
	}
	if a.userId != "" {
		page.UserId = a.userId
	}
	return a.client.Enqueue(page)
}

func (a *Object) identify() error {
	identify := analytics.Identify{
		AnonymousId:  a.config.AnonymousId,
		UserId:       a.userId,
	}
	return a.client.Enqueue(identify)
}

func (a *Object) malformedCommandError(e error) error {
	properties := make(analytics.Properties)
	properties.Set("error_message", e.Error())
	track := analytics.Track{
		AnonymousId: a.config.AnonymousId,
		Event:       "Malformed Command Error",
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

func (a *Object) addFlagProperties(cmd *cobra.Command) {
	flags := make(map[string]string)
	cmd.Flags().Visit(func(f *pflag.Flag){
		if contains(secretFlags, f.Name) {
			flags[f.Name] = secretValueString
		} else {
			flags[f.Name] = f.Value.String()
		}
	})
	a.properties.Set("flags", flags)
}

func  (a *Object) addArgsProperties(cmd *cobra.Command, args []string) {
	argsCopy := make([]string, len(args))
	if len(args) > 0 {
		if ids, ok := secretCommands[cmd.CommandPath()]; ok {
			copy(argsCopy, args)
			for _, i := range ids {
				argsCopy[i] = secretValueString
			}
		} else {
			copy(argsCopy, args)
		}
	}
	a.properties.Set("args", argsCopy)
}

func (a *Object) setUserInfo() {
	if a.config.CLIName == "ccloud" {
		cloudUserId, organizationId, email := a.getCloudUserInfo()
		if cloudUserId != "" {
			a.properties.Set("organization_id", organizationId)
			a.properties.Set("email", email)
		}
		a.userId = cloudUserId
	} else {
		a.userId = a.getCPUsername()
	}
}

func (a *Object) getCloudUserInfo() (userId string, organizationId string, email string) {
	if err := a.config.CheckLogin(); err != nil {
		return "", "", ""
	}
	user := a.config.Auth.User
	userId = strconv.Itoa(int(user.Id))
	organizationId = strconv.Itoa(int(user.OrganizationId))
	email = user.Email
	return userId, organizationId, email
}

func (a *Object) getCPUsername() string {
	if err := a.config.CheckLogin(); err != nil {
		return ""
	}
	ctx := a.config.Contexts[a.config.CurrentContext]
	cred := a.config.Credentials[ctx.Credential]
	return cred.Username
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if s == e {
			return true
		}
	}
	return false
}
