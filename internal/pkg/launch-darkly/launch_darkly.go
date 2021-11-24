package launch_darkly

import (
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/dghubble/sling"
)

const baseURL = "https://clientsdk.launchdarkly.com/sdk/eval/%s/users/%s"
const LD_PRODUCTION_ENV_ID = 5c636508aa445d32c86f26b1
stag = 5c63651f1df21432a45fc773
devel = 5c63653912b6db32db950445
cpd = 5c6365ffaa445d32c86f26c0
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