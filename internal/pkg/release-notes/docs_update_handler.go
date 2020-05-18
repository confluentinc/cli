package release_notes

import (
	"io/ioutil"
	"runtime"
	"strings"
)

const (
	ccloudDocsPageHeader = `.. _ccloud-release-notes:

==========================
|ccloud| CLI Release Notes
==========================
`
	confluentDocsPageHeader = `.. _cli-release-notes:

=============================
|confluent-cli| Release Notes
=============================
`
)

type DocsUpdateHandler interface {
	getUpdatedDocsPage(newReleaseNotes string) (string, error)
}

func NewDocsUpdateHandler(header string, docsFilePath string) DocsUpdateHandler {
	if runtime.GOOS == "windows" {
		header = strings.ReplaceAll(header, "\n", "\r\n")
	}
	return &DocsUpdateHandlerImp{
		pageHeader:   header,
		docsFilePath: docsFilePath,
	}
}

type DocsUpdateHandlerImp struct {
	pageHeader   string
	docsFilePath string
}

func (h *DocsUpdateHandlerImp) getUpdatedDocsPage(newReleaseNotes string) (string, error) {
	currentDocsPage, err := h.getCurrentDocsPage()
	if err != nil {
		return "", err
	}
	return h.addNewReleaseNotesToDocsPage(currentDocsPage, newReleaseNotes), nil
}

func (h *DocsUpdateHandlerImp) getCurrentDocsPage() (string, error) {
	b, err := ioutil.ReadFile(h.docsFilePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *DocsUpdateHandlerImp) addNewReleaseNotesToDocsPage(currentDocsPage string, newReleaseNotes string) string {
	previousReleaseNotes := strings.ReplaceAll(currentDocsPage, h.pageHeader, "")
	return h.pageHeader + "\n" + newReleaseNotes + "\n" + previousReleaseNotes
}
