package release_notes

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
)

const (
	releaseNotesLocalFileFormat  = "./release-notes/%s/%s"
	latestReleaseNotesFileName   = "latest-release.rst"
	releaseNotesDocsPageFileName = "release-notes.rst"

	prepFileName = "./release-notes/prep"
	placeHolder = "<PLACEHOLDER>"

	newFeaturesSectionTitle = "New Features"
	bugFixesSectionTitle = "Bug Fixes"

	noChangeContentFormat = "No change relating to %s CLI for this version."

	ccloudReleaseNotesFormat = `
%s Release Notes
==========================
%s`
	confluentReleaseNotesFormat = `
%s Release Notes
==========================
%s`

	confluentHeader = `.. _cli-release-notes:

=============================
|confluent-cli| Release Notes
=============================
`

	ccloudHeader = `.. _ccloud-release-notes:

==========================
|ccloud| CLI Release Notes
==========================
`
)
