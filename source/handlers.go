package source

import (
	"github.com/lectio/graph/model"
	"github.com/lectio/graph/observe"
)

// LinksAPIHandlerParams defines the parameters for the LinksAPIHandler function
type LinksAPIHandlerParams interface {
	Source() model.APISource
	ContentSettings() *model.ContentSettings
	HTTPClientSettings() *model.HTTPClientSettings
	LinksManager() *LinksManager
	Asynch() bool
	ProgressReporter() observe.ProgressReporter
}

// LinksAPIHandlerFunc is a function interface for any API ContentSource instances that return links
type LinksAPIHandlerFunc func(LinksAPIHandlerParams) (*model.Bookmarks, error)

type defaultLinksAPIHandlerParams struct {
	source           model.APISource
	hcs              *model.HTTPClientSettings
	lm               *LinksManager
	cs               *model.ContentSettings
	progressReporter observe.ProgressReporter
}

// NewLinksAPIHandlerParams returns a default parameters for common use cases
func NewLinksAPIHandlerParams(config *model.Configuration, source model.APISource, path model.SettingsPath) (LinksAPIHandlerParams, error) {
	result := new(defaultLinksAPIHandlerParams)

	result.source = source
	result.hcs = config.HTTPClientSettings(path)

	lls := config.LinkLifecyleSettings(path)
	httpClient := config.HTTPClient(path)
	result.lm = &LinksManager{Config: config, LinkSettings: lls, Client: httpClient}

	result.cs = config.ContentSettings(path)
	result.progressReporter = config.ProgressReporter

	return result, nil
}

func (p defaultLinksAPIHandlerParams) Source() model.APISource {
	return p.source
}

func (p defaultLinksAPIHandlerParams) ContentSettings() *model.ContentSettings {
	return p.cs
}

func (p defaultLinksAPIHandlerParams) HTTPClientSettings() *model.HTTPClientSettings {
	return p.hcs
}

func (p defaultLinksAPIHandlerParams) LinksManager() *LinksManager {
	return p.lm
}

func (p defaultLinksAPIHandlerParams) Asynch() bool {
	// link traversals can be slow so do it asynchronously; if we're not traversing links no need for extra work
	return p.lm.LinkSettings.TraverseLinks
}

func (p defaultLinksAPIHandlerParams) ProgressReporter() observe.ProgressReporter {
	return p.progressReporter
}
