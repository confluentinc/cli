package test

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	test_server "github.com/confluentinc/cli/test/test-server"

	"github.com/chromedp/chromedp"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	urlPlaceHolder     = "<URL_PLACEHOLDER>"
	savedToNetrcOutput = fmt.Sprintf(errors.WroteCredentialsToNetrcMsg, "/tmp/netrc_test")
	loggedInAsOutput   = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	loggedInEnvOutput  = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default")
)

func (s *CLITestSuite) TestCcloudLoginUseKafkaAuthKafkaErrors() {
	tests := []CLITest{
		{
			name:        "error if not authenticated",
			args:        "kafka topic create integ",
			fixture:     "err-not-authenticated.golden",
			wantErrCode: 1,
		},
		{
			name:        "error if no active kafka",
			args:        "kafka topic create integ",
			fixture:     "err-no-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
		//{
		//	name:        "error if topic already exists",
		//	args:        "kafka topic create topic-exist",
		//	fixture:     "topic-exists.golden",
		//	wantErrCode: 1,
		//	login:       "default",
		//	useKafka:    "lkc-create-topic",
		//	authKafka:   "true",
		//},
		{
			name:        "error if no api key used",
			args:        "kafka topic produce integ",
			fixture:     "err-no-api-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
		},
		{
			name:        "error if deleting non-existent api-key",
			args:        "api-key delete UNKNOWN",
			fixture:     "delete-unknown-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
			authKafka:   "true",
		},
		{
			name:        "error if using unknown kafka",
			args:        "kafka cluster use lkc-unknown",
			fixture:     "err-use-unknown-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
	}

	for _, tt := range tests {
		s.runCcloudTest(tt)
	}
}

func serveCloudBackend(t *testing.T) *test_server.TestBackend {
	router := test_server.NewCloudRouter(t)
	return test_server.NewCloudTestBackendFromRouters(router, test_server.NewEmptyKafkaRouter())
}

func serveMDSBackend(t *testing.T) *test_server.TestBackend {
	router := test_server.NewMdsRouter(t)
	return test_server.NewConfluentTestBackendFromRouter(router)
}

func (s *CLITestSuite) TestSaveUsernamePassword() {
	type saveTest struct {
		cliName  string
		want     string
		loginURL string
		bin      string
	}
	cloudBackend := serveCloudBackend(s.T())
	defer cloudBackend.Close()
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()
	tests := []saveTest{
		{
			"ccloud",
			"netrc-save-ccloud-username-password.golden",
			cloudBackend.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			"confluent",
			"netrc-save-mds-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	netrcInput := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc")
	for _, tt := range tests {
		// store existing credentials in netrc to check that they are not corrupted
		originalNetrc, err := ioutil.ReadFile(netrcInput)
		s.NoError(err)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		//TODO add save test using stdin input
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		s.Contains(output, loggedInAsOutput)
		if tt.cliName == "ccloud" {
			s.Contains(output, loggedInEnvOutput)
		}

		// check netrc file result
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(want), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestUpdateNetrcPassword() {
	type updateTest struct {
		input    string
		cliName  string
		want     string
		loginURL string
		bin      string
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	cloudServer := serveCloudBackend(s.T())
	defer cloudServer.Close()
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()
	tests := []updateTest{
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-ccloud"),
			"ccloud",
			"netrc-save-ccloud-username-password.golden",
			cloudServer.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-mds"),
			"confluent",
			"netrc-save-mds-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	for _, tt := range tests {
		// store existing credential + the user credential to be updated
		originalNetrc, err := ioutil.ReadFile(tt.input)
		s.NoError(err)
		originalNetrcString := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, []byte(originalNetrcString), 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		s.Contains(output, loggedInAsOutput)
		if tt.cliName == "ccloud" {
			s.Contains(output, loggedInEnvOutput)
		}

		// check netrc file result
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(want), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestSSOLoginAndSave() {
	if *skipSsoBrowserTests {
		s.T().Skip()
	}

	resetConfiguration(s.T(), "ccloud")

	err := ioutil.WriteFile(netrc.NetrcIntegrationTestFile, []byte{}, 0600)
	if err != nil {
		s.Fail("Failed to create netrc file")
	}

	env := []string{auth.CCloudEmailDeprecatedEnvVar + "=" + ssoTestEmail}
	cmd := exec.Command(binaryPath(s.T(), ccloudTestBin), []string{"login", "--save", "-vvv", "--url", ssoTestLoginUrl, "--no-browser"}...)
	cmd.Env = append(os.Environ(), env...)

	cliStdOut, err := cmd.StdoutPipe()
	s.NoError(err)
	cliStdIn, err := cmd.StdinPipe()
	s.NoError(err)
	cliStdErr, err := cmd.StderrPipe()
	s.NoError(err)

	scanner := bufio.NewScanner(cliStdOut)
	scannerErr := bufio.NewScanner(cliStdErr)
	go func() {
		var url string
		for scanner.Scan() {
			txt := scanner.Text()
			fmt.Println("CLI output | " + txt)
			if url == "" {
				url = parseSsoAuthUrlFromOutput([]byte(txt))
			}
			if strings.Contains(txt, "paste the code here") {
				break
			}
		}

		if url == "" {
			s.Fail("CLI did not output auth URL")
		} else {
			token := s.ssoAuthenticateViaBrowser(url)
			_, e := cliStdIn.Write([]byte(token))
			s.NoError(e)
			e = cliStdIn.Close()
			s.NoError(e)
			printedLoginMessage := false
			for scannerErr.Scan() {
				if strings.Contains(scannerErr.Text()+"\n", fmt.Sprintf(errors.LoggedInAsMsg, ssoTestEmail)) {
					printedLoginMessage = true
					break
				}
			}
			s.True(printedLoginMessage)

		}
	}()

	err = cmd.Start()
	s.NoError(err)

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	timeout := time.After(60 * time.Second)

	select {
	case <-timeout:
		s.Fail("Timed out. The CLI may have printed out something unexpected or something went awry in the okta browser auth flow.")
	case err := <-done:
		// the output from the cmd.Wait(). Should not have an error status
		s.NoError(err)
	}

	// Verifying login --save functionality by checking netrc file
	got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
	s.NoError(err)
	pattern := `machine\sconfluent-cli:ccloud-sso-refresh-token:login-ziru\+paas-integ-sso@confluent.io-https://devel.cpdev.cloud\r?\n\s+login\sziru\+paas-integ-sso@confluent.io\r?\n\s+password\s[\w-]+`
	match, err := regexp.Match(pattern, got)
	s.NoError(err)
	if !match {
		fmt.Println("Refresh token credential not written to netrc file properly.")
		want := "machine confluent-cli:ccloud-sso-refresh-token:login-ziru+paas-integ-sso@confluent.io-https://devel.cpdev.cloud\n	login ziru+paas-integ-sso@confluent.io\n	password <refresh_token>"
		msg := fmt.Sprintf("expected: %s\nactual: %s\n", want, got)
		s.Fail("sso login command with --save flag failed to properly write refresh token credential.\n" + msg)
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}


func parseSsoAuthUrlFromOutput(output []byte) string {
	regex, err := regexp.Compile(`.*([\S]*connection=` + ssoTestConnectionName + `).*`)
	if err != nil {
		panic("Error compiling regex")
	}
	groups := regex.FindSubmatch(output)
	if groups == nil || len(groups) < 2 {
		return ""
	}
	authUrl := string(groups[0])
	return authUrl
}

func (s *CLITestSuite) ssoAuthenticateViaBrowser(authUrl string) string {
	opts := append(chromedp.DefaultExecAllocatorOptions[:]) // uncomment to disable headless mode and see the actual browser
	//chromedp.Flag("headless", false),

	var err error
	var taskCtx context.Context
	tries := 0
	for tries < 5 {
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()
		taskCtx, cancel = chromedp.NewContext(allocCtx)
		defer cancel()
		// ensure that the browser process is started
		if err = chromedp.Run(taskCtx); err != nil {
			fmt.Println("Caught error when starting chrome. Will retry. Error was: " + err.Error())
			tries += 1
		} else {
			fmt.Println("Successfully started chrome")
			break
		}
	}
	if err != nil {
		s.NoError(err, fmt.Sprintf("Could not start chrome after %d tries. Error was: %s\n", tries, err))
	}

	// navigate to authUrl
	fmt.Println("Navigating to authUrl...")
	err = chromedp.Run(taskCtx, chromedp.Navigate(authUrl))
	s.NoError(err)
	fmt.Println("Inputing credentials to Okta...")
	err = chromedp.Run(taskCtx, chromedp.WaitVisible(`//input[@name="username"]`))
	s.NoError(err)
	err = chromedp.Run(taskCtx, chromedp.SendKeys(`//input[@id="okta-signin-username"]`, ssoTestEmail))
	s.NoError(err)
	err = chromedp.Run(taskCtx, chromedp.SendKeys(`//input[@id="okta-signin-password"]`, ssoTestPassword))
	s.NoError(err)
	fmt.Println("Submitting login request to Okta..")
	err = chromedp.Run(taskCtx, chromedp.Click(`//input[@id="okta-signin-submit"]`))
	s.NoError(err)
	fmt.Println("Waiting for CCloud to load...")
	err = chromedp.Run(taskCtx, chromedp.WaitVisible(`//div[@id="cc-root"]`))
	s.NoError(err)
	fmt.Println("CCloud is loaded, grabbing auth token...")
	var token string
	// chromedp waits until it finds the element on the page. If there's some error and the element
	// does not load correctly, this will wait forever and the test will time out
	// There's not a good workaround for this, but to debug, it's helpful to disable headless mode (commented above)
	err = chromedp.Run(taskCtx, chromedp.Text(`//div[@id="token"]`, &token))
	s.NoError(err)
	fmt.Println("Successfully logged in and retrieved auth token")
	return token
}

func (s *CLITestSuite) TestMDSLoginURL() {
	tests := []CLITest{
		{
			name:        "invalid URL provided",
			args:        "login --url http:///test",
			fixture:     "invalid-login-url.golden",
			wantErrCode: 1,
		},
	}
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()

	for _, tt := range tests {
		tt.loginURL = mdsServer.GetMdsUrl()
		s.runConfluentTest(tt)
	}
}
