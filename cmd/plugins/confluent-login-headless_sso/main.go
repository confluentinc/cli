package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/plugin"
)

type state struct {
	provider string
	email    string
	password string
	code     string
}

func main() {
	cmd := cobra.Command{
		Use:     "headless-sso",
		Short:   "Automatically authenticate to Confluent Cloud with SSO.",
		Long:    "Use a headless browser to automatically authenticate to Confluent Cloud with SSO. Best suited for CI jobs which cannot rely on human interaction.",
		RunE:    login,
		Example: "confluent login headless-sso --provider okta --email example@confluent.io --password $(cat password.txt)",
	}

	cmd.Flags().String("provider", "", "SSO provider. Supported providers: okta")
	cmd.Flags().String("email", "", "Confluent Cloud SSO email.")
	cmd.Flags().String("password", "", "SSO password.")
	cmd.Flags().String("url", "", "Confluent Cloud URL.")

	cobra.CheckErr(cmd.MarkFlagRequired("provider"))
	cobra.CheckErr(cmd.MarkFlagRequired("email"))
	cobra.CheckErr(cmd.MarkFlagRequired("password"))

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func login(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true
	cmd.SetOut(os.Stdout)

	provider, err := cmd.Flags().GetString("provider")
	cobra.CheckErr(err)

	email, err := cmd.Flags().GetString("email")
	cobra.CheckErr(err)

	password, err := cmd.Flags().GetString("password")
	cobra.CheckErr(err)

	url, err := cmd.Flags().GetString("url")
	cobra.CheckErr(err)

	var (
		stdin  = make(chan string, 1)
		stdout = make(chan string)
		stderr = make(chan string)
		done   = make(chan bool)
	)

	go func() {
		args := []string{"login", "--no-browser"}
		if url != "" {
			args = append(args, "--url", url)
		}
		plugin.Capture(exec.Command("confluent", args...), stdin, stdout, stderr)
		done <- true
	}()

	s := state{
		provider: provider,
		email:    email,
		password: password,
	}

	line := ""
	for {
		select {
		case out := <-stdout:
			line += out

			cmd.Print(out)
			if err := s.handle(line, stdin); err != nil {
				return err
			}

			if out == "\n" {
				line = ""
			}
		case err := <-stderr:
			cmd.PrintErr(err)
		case <-done:
			return nil
		}
	}
}

// handle specific lines of output, and simulate the user by producing to stdin.
func (s *state) handle(line string, stdin chan string) error {
	if line == "Email: " {
		stdin <- s.email + "\n"
	}

	if strings.HasSuffix(line, "Password: ") {
		return fmt.Errorf("non-SSO user")
	}

	if strings.HasPrefix(line, "https://") && strings.HasSuffix(line, "\n") {
		code, err := s.authenticate(line)
		if err != nil {
			return err
		}
		s.code = code
	}

	if line == "After authenticating in your browser, paste the code here:\n" {
		stdin <- s.code + "\n"
	}

	return nil
}

// authenticate with a headless browser at the provided URL and retrieve an authentication code.
func (s *state) authenticate(url string) (string, error) {
	switch s.provider {
	case "okta":
		return okta(url, s.email, s.password), nil
	default:
		return "", fmt.Errorf(`unsupported provider "%s"`, s.provider)
	}
}
