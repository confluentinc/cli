// golicense-downloader downloads LICENSE and NOTICE files for each dependency found by github.com/mitchellh/golicense
//
// Usage:
//    GITHUB_TOKEN=${token} golicense .golicense.hcl my-tool | GITHUB_TOKEN=${token} golicense-downloader -f .golicense-downloader.json -l legal/my-tool/licenses -n legal/my-tool/notices
package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v25/github"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

const (
	githubTokenEnvVar    = "GITHUB_TOKEN"
	licenseFilenameFmt   = "LICENSE-%s_%s.txt"
	noticeFilenameFmt    = "NOTICE-%s_%s.txt"
	licenseIndexFilename = "licenses.txt"
)

var (
	noticeFiles = []string{"NOTICE", "NOTICES", "NOTICE.txt", "NOTICES.txt"}
	licenseDir  = pflag.StringP("licenses-dir", "l", "./legal/licenses", "Directory in which to write licenses")
	noticeDir   = pflag.StringP("notices-dir", "n", "./legal/notices", "Directory in which to write notices")
	configFile  = pflag.StringP("config-file", "F", "", "File from which to read golicense-downloader configuration")
)

type Config struct {
	DepOverrides map[string]string `json:"depOverrides"`
}

type LicenseDownloader struct {
	*Config
	Client       *github.Client
	LicenseIndex string
	LicenseFmt   string
	NoticeFmt    string
}

// License represents a software LICENSE obtained from golicense / github
type License struct {
	Owner   string
	Repo    string
	Version string // TODO, this isn't provided by golicense. We assume latest license is correct.
	License string
}

func main() {
	// Parse and validate flags
	pflag.Parse()

	// Validate usage - pipe from golicense
	info, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprintf(os.Stderr, "Please pipe the output from golicense")
		os.Exit(1)
	}

	// Validate usage - set GITHUB_TOKEN env var
	token, ok := os.LookupEnv(githubTokenEnvVar)
	if !ok || token == "" {
		fmt.Fprintf(os.Stderr, "Missing environment variable: %s\n", githubTokenEnvVar)
		os.Exit(1)
	}

	// Load config file, if any
	config := &Config{}
	if *configFile != "" {
		b, err := ioutil.ReadFile(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read config file at %s: %s", *configFile, err)
			os.Exit(1)
		}
		err = json.Unmarshal(b, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to unmarshal config from %s: %s", *configFile, err)
			os.Exit(1)
		}
	}

	// Instantiate LicenseDownloader
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	downloader := &LicenseDownloader{
		Config:       config,
		Client:       client,
		LicenseIndex: filepath.Join(filepath.Dir(*licenseDir), licenseIndexFilename),
		LicenseFmt:   filepath.Join(*licenseDir, licenseFilenameFmt),
		NoticeFmt:    filepath.Join(*noticeDir, noticeFilenameFmt),
	}

	// Run it!
	if err = downloader.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func (g *LicenseDownloader) Run(ctx context.Context) error {
	// We both remove any existing files and create any missing parent directories for licenses and notices
	licenseDir := filepath.Dir(fmt.Sprintf(g.LicenseFmt, "example", "example"))
	err := os.RemoveAll(licenseDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(licenseDir, os.ModePerm)
	if err != nil {
		return err
	}
	noticeDir := filepath.Dir(fmt.Sprintf(g.NoticeFmt, "example", "example"))
	err = os.RemoveAll(noticeDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(noticeDir, os.ModePerm)
	if err != nil {
		return err
	}

	var licenses []*License
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		license, err := g.ParseLicense(text)
		if err != nil {
			return err
		}
		if license == nil {
			continue
		}
		licenses = append(licenses, license)

		// Download the project LICENSE text from github
		err = g.DownloadLicense(ctx, license.Owner, license.Repo)
		if err != nil {
			return err
		}

		// Download the project NOTICE text from github
		err = g.DownloadNotice(ctx, license.Owner, license.Repo)
		if err != nil {
			return err
		}
	}

	// Create licenses.txt index
	return g.CreateLicenseIndex(licenses)
}

func (g *LicenseDownloader) CreateLicenseIndex(licenses []*License) error {
	indexFile, err := os.Create(g.LicenseIndex)
	if err != nil {
		return err
	}
	writer := tablewriter.NewWriter(indexFile)
	writer.SetAutoWrapText(false)
	writer.SetHeader([]string{"Artifact", "License"})
	for _, license := range licenses {
		writer.Append([]string{fmt.Sprintf("github.com/%s/%s", license.Owner, license.Repo), license.License})
	}
	writer.Render()
	return nil
}

func (g *LicenseDownloader) ParseLicense(text string) (*License, error) {
	text = strings.Replace(text, "\n", "", -1) // convert CRLF to LF
	columns := strings.SplitN(text, " ", 2)
	if len(columns) != 2 {
		return nil, fmt.Errorf("invalid golicense output: %s\n", text)
	}
	dep, license := strings.TrimSpace(columns[0]), strings.TrimSpace(columns[1])
	if override, ok := g.DepOverrides[dep]; ok {
		dep = override
	}
	if !strings.HasPrefix(dep, "github.com") {
		// ignore golang stdlib sub-repo packages
		if !strings.HasPrefix(dep, "golang.org/x/") {
			fmt.Fprintf(os.Stderr, "Unable to fetch license for %s\n", dep)
		}
		return nil, nil
	}

	parts := strings.Split(dep, "/")

	if len(parts) > 3 {
		fmt.Printf("Possible sub-package referenced by go.sum, with github url: %s ; make sure the parent package is in go.sum\n", dep)
	} else if len(parts) < 3 {
		return nil, fmt.Errorf("invalid github url: %s", dep)
	}
	owner, repo := parts[1], parts[2]
	return &License{Owner: owner, Repo: repo, License: license}, nil
}

func (g *LicenseDownloader) DownloadLicense(ctx context.Context, owner, repo string) error {
	license, err := g.GetLicense(ctx, owner, repo)
	if err != nil {
		return err
	}
	if license == "" {
		// TODO: HACK: This is because of the overrides documented in .golicense.hcl
		if owner != "confluentinc" {
			fmt.Fprintf(os.Stderr, "No contents found for github.com/%s/%s\n", owner, repo)
		}
		return nil
	}
	return ioutil.WriteFile(fmt.Sprintf(g.LicenseFmt, owner, repo), []byte(license), os.ModePerm)
}

func (g *LicenseDownloader) DownloadNotice(ctx context.Context, owner, repo string) error {
	notice, err := g.GetNotice(ctx, owner, repo)
	if err != nil {
		return err
	}
	if notice == "" {
		return nil
	}
	return ioutil.WriteFile(fmt.Sprintf(g.NoticeFmt, owner, repo), []byte(notice), os.ModePerm)
}

func (g *LicenseDownloader) GetLicense(ctx context.Context, owner, repo string) (string, error) {
	lic, resp, err := g.Client.Repositories.License(ctx, owner, repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", err
	}
	contents, err := base64.StdEncoding.DecodeString(*lic.Content)
	return string(contents), err
}

func (g *LicenseDownloader) GetNotice(ctx context.Context, owner, repo string) (string, error) {
	for _, noticeFile := range noticeFiles {
		notice, _, resp, err := g.Client.Repositories.GetContents(ctx, owner, repo, noticeFile,
			&github.RepositoryContentGetOptions{Ref: "master"})
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return "", nil
			}
			return "", err
		}
		return notice.GetContent()
	}
	return "", nil
}
