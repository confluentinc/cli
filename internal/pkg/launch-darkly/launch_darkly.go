package launch_darkly

import (
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/dghubble/sling"
)

const baseURL = "https://clientsdk.launchdarkly.com/sdk/eval/%s/users/%s"
const LD_PRODUCTION_ENV_CLIENT_ID = "61af57740127630ce47de5be"
const LD_TEST_ENV_CLIENT_ID = "61af57740127630ce47de5bd"

type LaunchDarklyManager interface {
	EvaluateStringFlag(flagName string, defaultVal string) string
	EvaluateBoolFlag(flagName string, defaultVal bool) bool
}

type LaunchDarklyManagerImpl struct {
	logger       *log.Logger
	client       *sling.Sling
}

req, err := sling.New().Get(fmt.Sprintf("https://clientsdk.launchdarkly.com/sdk/eval/%s/users/%s", LD_PRODUCTION_ENV_ID, base64str)).Receive(&flags, err)

func NewLaunchDarklyManager(logger *log.Logger) LaunchDarklyManager {
	return &LaunchDarklyManagerImpl{
		logger: logger,
		client: sling.New().Base(baseURL),
	}
}

func (l *LaunchDarklyManagerImpl) EvaluateStringFlag(flagName string, defaultVal string) string {
	l.client.New()
}

func (l *LaunchDarklyManagerImpl) EvaluateBoolFlag(flagName string, defaultVal bool) bool {

}