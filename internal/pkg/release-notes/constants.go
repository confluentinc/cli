package release_notes

import "path"

type ReleaseNotesSection int

const (
	bothNewFeature ReleaseNotesSection = iota
	bothBugFix
	ccloudNewFeature
	ccloudBugFix
	confluentNewFeature
	confluentBugFix
)

var (
	sectionNameMap = map[string]ReleaseNotesSection{
		"New Features for Both CLIs": bothNewFeature,
		"Bug Fixes for Both CLIs": bothBugFix,
		"CCloud New Features": ccloudNewFeature,
		"CCloud Bug Fixes": ccloudBugFix,
		"Confluent New Features": confluentNewFeature,
		"Confluent Bug Fixes": confluentBugFix,
	}
	releaseNotesLocalFile = "./release-notes/%s/%s"
	latestReleaseNotesFileName = "latest-release.rst"
	fullReleaseNotesFileName = "index.rst"

	prepFileName = path.Join(".", "release-notes", "prep")
	placeHolder = "<PLACEHOLDER>"

	newFeaturesSectionTitle = "New Features"
	bugFixesSectionTitle = "Bug Fixes"

	//releaseNotesMsgFile = path.Join(".", "internal", "pkg", "update" , "update_msg.go")


	releaseNotesContentFormat = `
=============================================================
%s CLI %s Release Notes
=============================================================

%s
`
)
