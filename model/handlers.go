package model

import (
	"github.com/lectio/graph/observe"
)

// LinksAPIHandlerParams defines the parameters for the LinksAPIHandler function
type LinksAPIHandlerParams interface {
	Source() APISource
	ContentSettings() *ContentSettings
	HTTPClientSettings() *HTTPClientSettings
	LinkLifecyleSettings() *LinkLifecyleSettings
	Asynch() bool
	ProgressReporter() observe.ProgressReporter
}

// LinksAPIHandlerFunc is a function interface for any API ContentSource instances that return links
type LinksAPIHandlerFunc func(LinksAPIHandlerParams) (*Bookmarks, error)

type defaultLinksAPIHandlerParams struct {
	source           APISource
	hcs              *HTTPClientSettings
	lls              *LinkLifecyleSettings
	cs               *ContentSettings
	progressReporter observe.ProgressReporter
}

// NewLinksAPIHandlerParams returns a default parameters for common use cases
func NewLinksAPIHandlerParams(config *Configuration, source APISource, path SettingsPath) (LinksAPIHandlerParams, error) {
	result := new(defaultLinksAPIHandlerParams)

	result.source = source
	result.hcs = config.HTTPClientSettings(path)
	result.lls = config.LinkLifecyleSettings(path)
	result.cs = config.ContentSettings(path)
	result.progressReporter = config.ProgressReporter()

	return result, nil
}

func (p defaultLinksAPIHandlerParams) Source() APISource {
	return p.source
}

func (p defaultLinksAPIHandlerParams) ContentSettings() *ContentSettings {
	return p.cs
}

func (p defaultLinksAPIHandlerParams) HTTPClientSettings() *HTTPClientSettings {
	return p.hcs
}

func (p defaultLinksAPIHandlerParams) LinkLifecyleSettings() *LinkLifecyleSettings {
	return p.lls
}

func (p defaultLinksAPIHandlerParams) Asynch() bool {
	// link traversals can be slow so do it asynchronously; if we're not traversing links no need for extra work
	return p.lls.TraverseLinks
}

func (p defaultLinksAPIHandlerParams) ProgressReporter() observe.ProgressReporter {
	return p.progressReporter
}
