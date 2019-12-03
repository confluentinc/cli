package test

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	skip = flag.Bool("skip-sso-browser-tests", false, "If flag is preset, run the tests that require a web browser.")

	// browser tests by default against devel
	testEmail = *flag.String("sso-test-user-email", "ziru+paas-integ-sso@confluent.io", "The email of an sso enabled test user.")
	testPassword = *flag.String("sso-test-user-password","aWLw9eG+F", "The password for the sso enabled test user.")
	// this connection is preconfigured in Auth0 to hit a test Okta account
	testConnectionName = *flag.String("sso-test-connection-name", "confluent-dev", "The Auth0 SSO connection name.")
	loginURL = *flag.String("sso-test-login-url", "https://devel.cpdev.cloud","The login url to use.")
)


func (s *CLITestSuite) TestSSOLogin() {
	t := s.T()
	if *skip {
		t.Skip()
	}

	resetConfiguration(s.T(), "ccloud")

	env := []string{"XX_CCLOUD_EMAIL="+testEmail}

	binaryName := "ccloud"
	args := "login --url "+loginURL+" --no-browser"
	path := binaryPath(t, binaryName)
	_, _ = fmt.Println(path, args)
	cmd := exec.Command(path, strings.Split(args, " ")...)
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
			fmt.Println("CLI output | "+txt)
			if url == "" {
				url = parseAuthUrlFromOutput([]byte(txt))
			}
			if strings.Contains(txt, "paste the code here") {
				break
			}
		}

		if url == "" {
			s.Fail("CLI did not output auth URL")
		} else {
			token := authenticateViaBrowser(s, url)
			_, e := cliStdIn.Write([]byte(token))
			s.NoError(e)
			e = cliStdIn.Close()
			s.NoError(e)

			scanner.Scan()
			s.Equal(scanner.Text(), "Logged in as "+testEmail)
		}
	}()

	err = cmd.Start()
	s.NoError(err)

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	timeout := time.After(20 * time.Second)

	select {
	case <-timeout:
		s.Fail("Timed out. Either there was an error or the CLI printed out something unexpected.")
	case err := <-done:
		// the output from the cmd.Wait(). Should not have an error status
		s.NoError(err)
	}
}

func parseAuthUrlFromOutput(output []byte) string {
	regex, err := regexp.Compile(`.*([\S]*connection=`+testConnectionName+`).*`)
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

func authenticateViaBrowser(s *CLITestSuite, authUrl string) string {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// uncomment to disable headless mode and see the actual browser
		//chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		s.NoError(err)
	}
	// navigate to authUrl
	fmt.Println("Navigating to authUrl...")
	err := chromedp.Run(taskCtx, chromedp.Navigate(authUrl))
	s.NoError(err)
	fmt.Println("Inputing credentials to Okta...")
	err = chromedp.Run(taskCtx, chromedp.WaitVisible(`//input[@name="username"]`))
	s.NoError(err)
	err = chromedp.Run(taskCtx, chromedp.SendKeys(`//input[@id="okta-signin-username"]`, testEmail))
	s.NoError(err)
	err = chromedp.Run(taskCtx, chromedp.SendKeys(`//input[@id="okta-signin-password"]`, testPassword))
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
	// TODO update this with the element's unique id once the PR is merged in frontend vault
	err = chromedp.Run(taskCtx, chromedp.Text(`//div[@class="UniversalLogin__CodeBox-sc-10srvam-3 feAcoO"]`, &token))
	s.NoError(err)
	fmt.Println("Successfully logged in and retrieved auth token")
	return token
}

