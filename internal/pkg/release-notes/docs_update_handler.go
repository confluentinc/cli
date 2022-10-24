package releasenotes

import (
	"os"
	"strings"
)

const docsPageHeader = `.. _cli-release-notes:

=============================
|confluent-cli| Release Notes
=============================
`

type DocsUpdateHandler interface {
	getUpdatedDocsPage(newReleaseNotes string) (string, error)
}

func NewDocsUpdateHandler(header, docsFilePath string) DocsUpdateHandler {
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
	b, err := os.ReadFile(h.docsFilePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *DocsUpdateHandlerImp) addNewReleaseNotesToDocsPage(currentDocsPage, newReleaseNotes string) string {
	previousReleaseNotes := strings.ReplaceAll(currentDocsPage, h.pageHeader, "")
	return h.pageHeader + "\n" + newReleaseNotes + "\n" + previousReleaseNotes
}
