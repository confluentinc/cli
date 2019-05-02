package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/google/go-github/v25/github"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
)

const (
	netrcMachine = "api.github.com"
	licenseFmt   = "legal/licenses/LICENSE-%s-%s.txt"
	noticeFmt    = "legal/notices/NOTICE-%s-%s.txt"
)

var (
	noticeFiles = []string{"NOTICE", "NOTICES", "NOTICE.txt", "NOTICES.txt"}
)

func main() {
	ctx := context.Background()

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		fmt.Println("Please pipe the output from golicense")
		return
	}

	token, err := getAuthFromNetrc(netrcMachine)
	if err != nil {
		panic(err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	licenseDir := filepath.Dir(fmt.Sprintf(licenseFmt, "example", "example"))
	err = os.MkdirAll(licenseDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	noticeDir := filepath.Dir(fmt.Sprintf(noticeFmt, "example", "example"))
	err = os.MkdirAll(noticeDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		text = strings.Replace(text, "\n", "", -1) // convert CRLF to LF
		dep := strings.Split(text, " ")[0]
		if !strings.HasPrefix(dep, "github.com") {
			fmt.Fprintf(os.Stderr, "Unable to fetch license for %s\n", dep)
			continue
		}

		parts := strings.Split(dep, "/")
		if len(parts) != 3 {
			panic(fmt.Errorf("invalid github url: %s", dep))
		}
		owner, repo := parts[1], parts[2]

		// Get the LICENSE
		lic, resp, err := client.Repositories.License(ctx, owner, repo)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				// This is because of the new unlicensed repos documented in .golicense.hcl
				if owner != "confluentinc" {
					fmt.Fprintf(os.Stderr, "No license found for %s\n", dep)
				}
				continue
			}
			panic(err)
		}
		contents, err := base64.StdEncoding.DecodeString(*lic.Content)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(fmt.Sprintf(licenseFmt, owner, repo), contents, os.ModePerm)
		if err != nil {
			panic(err)
		}

		// Get the NOTICE
		for _, noticeFile := range noticeFiles {
			notice, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, noticeFile,
				&github.RepositoryContentGetOptions{Ref: "master"})
			if err != nil {
				if resp.StatusCode == http.StatusNotFound {
					continue
				}
				panic(err)
			}
			contents, err := notice.GetContent()
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(fmt.Sprintf(licenseFmt, owner, repo), []byte(contents), os.ModePerm)
			if err != nil {
				panic(err)
			}
			break
		}
	}
}

// Borrowed from https://github.com/hashicorp/go-getter/blob/master/netrc.go
func getAuthFromNetrc(host string) (string, error) {
	// Get the netrc file path
	path := os.Getenv("NETRC")
	if path == "" {
		filename := ".netrc"
		if runtime.GOOS == "windows" {
			filename = "_netrc"
		}

		var err error
		path, err = homedir.Expand("~/" + filename)
		if err != nil {
			return "", err
		}
	}

	// If the file is not a file, then do nothing
	if fi, err := os.Stat(path); err != nil {
		// File doesn't exist, do nothing
		if os.IsNotExist(err) {
			return "", nil
		}

		// Some other error!
		return "", err
	} else if fi.IsDir() {
		// File is directory, ignore
		return "", nil
	}

	// Load up the netrc file
	net, err := netrc.ParseFile(path)
	if err != nil {
		return "", fmt.Errorf("error parsing netrc file at %q: %s", path, err)
	}

	machine := net.FindMachine(host)
	if machine == nil {
		// Machine not found, no problem
		return "", nil
	}

	return machine.Password, nil
}
