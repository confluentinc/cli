package test

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/confluentinc/cli/internal/pkg/auth"
)

var (
	urlPlaceHolder       = "<URL_PLACEHOLDER>"
	ccloudLoginOutput    = "Written credentials to file /tmp/netrc_test\nLogged in as good@user.com\nUsing environment a-595 (\"default\")\n"
	confluentLoginOutput = "Written credentials to file /tmp/netrc_test\nLogged in as good@user.com\n"
)

func (s *CLITestSuite) Test_Ccloud_Login_UseKafka_AuthKafka_Errors() {
	tests := []CLITest{
		{
			name:    "error if not authenticated",
			args:    "kafka topic create integ",
			fixture: "err-not-authenticated.golden",
		},
		{
			name:    "error if no active kafka",
			args:    "kafka topic create integ",
			fixture: "err-no-kafka.golden",
			login:   "default",
		},
		{
			name:      "error if topic already exists",
			args:      "kafka topic create integ",
			fixture:   "topic-exists.golden",
			login:     "default",
			useKafka:  "lkc-abc123",
			authKafka: "true",
		},
		{
			name:     "error if no api key used",
			args:     "kafka topic produce integ",
			fixture:  "err-no-api-key.golden",
			login:    "default",
			useKafka: "lkc-abc123",
		},
		{
			name:      "error if deleting non-existent api-key",
			args:      "api-key delete UNKNOWN",
			fixture:   "delete-unknown-key.golden",
			login:     "default",
			useKafka:  "lkc-abc123",
			authKafka: "true",
		},
		{
			name:    "error if using unknown kafka",
			args:    "kafka cluster use lkc-unknown",
			fixture: "err-use-unknown-kafka.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		if strings.HasPrefix(tt.name, "error") {
			tt.wantErrCode = 1
		}
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}

func serveLogin(t *testing.T) *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/sessions", handleLogin(t))
	router.HandleFunc("/api/check_email/", handleCheckEmail(t))
	router.HandleFunc("/api/me", handleMe(t))
	return httptest.NewServer(router)
}

func (s *CLITestSuite) Test_Save_Username_Password() {
	t := s.T()
	type saveTest struct {
		cliName  string
		want     string
		loginURL string
	}
	tests := []saveTest{
		{
			"ccloud",
			"netrc-save-ccloud-username-password.golden",
			serveLogin(t).URL,
		},
		{
			"confluent",
			"netrc-save-mds-username-password.golden",
			serveMds(t).URL,
		},
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}
	netrcInput := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc")
	for _, tt := range tests {
		// store existing credentials in netrc to check that they are not corrupted
		originalNetrc, err := ioutil.ReadFile(netrcInput)
		s.NoError(err)
		err = ioutil.WriteFile(auth.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var bin string
		var expectedOutput string
		if tt.cliName == "ccloud" {
			bin = ccloudTestBin
			expectedOutput = ccloudLoginOutput
		} else {
			bin = confluentTestBin
			expectedOutput = confluentLoginOutput
		}
		env := []string{"XX_CCLOUD_EMAIL=good@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(t, bin, env, "login --save --url "+tt.loginURL, 0)
		s.Equal(expectedOutput, output)

		// check netrc file result
		got, err := ioutil.ReadFile(auth.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(NormalizeNewLines(want), NormalizeNewLines(string(got)))
	}
	_ = os.Remove(auth.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) Test_Update_Netrc_Password() {
	t := s.T()
	type updateTest struct {
		input    string
		cliName  string
		want     string
		loginURL string
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}
	tests := []updateTest{
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-ccloud"),
			"ccloud",
			"netrc-save-ccloud-username-password.golden",
			serveLogin(t).URL,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-mds"),
			"confluent",
			"netrc-save-mds-username-password.golden",
			serveMds(t).URL,
		},
	}
	for _, tt := range tests {
		// store existing credential + the user credential to be updated
		originalNetrc, err := ioutil.ReadFile(tt.input)
		s.NoError(err)
		originalNetrcString := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = ioutil.WriteFile(auth.NetrcIntegrationTestFile, []byte(originalNetrcString), 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var bin string
		var expectedOutput string
		if tt.cliName == "ccloud" {
			bin = ccloudTestBin
			expectedOutput = ccloudLoginOutput
		} else {
			bin = confluentTestBin
			expectedOutput = confluentLoginOutput
		}
		env := []string{"XX_CCLOUD_EMAIL=good@user.com", "XX_CCLOUD_PASSWORD=pass1"}
		output := runCommand(t, bin, env, "login --save --url "+tt.loginURL, 0)
		s.Equal(expectedOutput, output)

		// check netrc file result
		got, err := ioutil.ReadFile(auth.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(NormalizeNewLines(want), NormalizeNewLines(string(got)))
	}
	_ = os.Remove(auth.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) Test_SSO_Login_And_Save() {
	t := s.T()
	if *skipSsoBrowserTests {
		t.Skip()
	}

	resetConfiguration(s.T(), "ccloud")

	env := []string{"XX_CCLOUD_EMAIL=" + ssoTestEmail}
	cmd := exec.Command(binaryPath(t, ccloudTestBin), []string{"login", "--save", "--url", ssoTestLoginUrl, "--no-browser"}...)
	cmd.Env = append(os.Environ(), env...)

	cliStdOut, err := cmd.StdoutPipe()
	s.NoError(err)
	cliStdIn, err := cmd.StdinPipe()
	s.NoError(err)

	scanner := bufio.NewScanner(cliStdOut)
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

			scanner.Scan()
			s.Equal("Logged in as "+ssoTestEmail, scanner.Text())
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
	got, err := ioutil.ReadFile(auth.NetrcIntegrationTestFile)
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
	_ = os.Remove(auth.NetrcIntegrationTestFile)
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
