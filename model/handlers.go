package model

import (
	"github.com/lectio/graph/observe"
)

// LinksAPIHandlerParams defines the parameters for the LinksAPIHandler function
type LinksAPIHandlerParams interface {
	Source() APISource
	Settings() *SettingsBundle
	Asynch() bool
	ProgressReporter() observe.ProgressReporter
}

// LinksAPIHandlerFunc is a function interface for any API ContentSource instances that return links
type LinksAPIHandlerFunc func(LinksAPIHandlerParams) (*Bookmarks, error)

type defaultLinksAPIHandlerParams struct {
	source             APISource
	settingsBundleName SettingsBundleName
	settings           *SettingsBundle
	progressReporter   observe.ProgressReporter
}

// NewLinksAPIHandlerParams returns a default parameters for common use cases
func NewLinksAPIHandlerParams(config *Configuration, source APISource, settings *SettingsBundle) (LinksAPIHandlerParams, error) {
	result := new(defaultLinksAPIHandlerParams)
	result.source = source
	result.settings = settings
	result.progressReporter = result.settings.Observe.ProgressReporter()

	return result, nil
}

func (p defaultLinksAPIHandlerParams) Source() APISource {
	return p.source
}

func (p defaultLinksAPIHandlerParams) Settings() *SettingsBundle {
	return p.settings
}

func (p defaultLinksAPIHandlerParams) Asynch() bool {
	// link traversals can be slow so do it asynchronously; if we're not traversing links no need for extra work
	return p.settings.Links.TraverseLinks
}

func (p defaultLinksAPIHandlerParams) ProgressReporter() observe.ProgressReporter {
	return p.progressReporter
}
