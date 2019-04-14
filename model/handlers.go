package model

import (
	"fmt"

	"github.com/lectio/graph/observe"

	lc "github.com/lectio/link/cache"
)

// LinksAPIHandlerParams defines the parameters for the LinksAPIHandler function
type LinksAPIHandlerParams interface {
	Source() *APISource
	Settings() *SettingsBundle
	LinksCache() lc.Cache
	ProgressReporter() observe.ProgressReporter
}

// LinksAPIHandlerFunc is a function interface for any API ContentSource instances that return links
type LinksAPIHandlerFunc func(LinksAPIHandlerParams) (*HarvestedLinks, error)

type defaultLinksAPIHandlerParams struct {
	source             *APISource
	settingsBundleName SettingsBundleName
	settings           *SettingsBundle
	linksCache         lc.Cache
	progressReporter   observe.ProgressReporter
}

// NewLinksAPIHandlerParams returns a default parameters for common use cases
func NewLinksAPIHandlerParams(config *Configuration, source *APISource, settingsBundleName SettingsBundleName) (LinksAPIHandlerParams, error) {
	result := new(defaultLinksAPIHandlerParams)
	result.source = source
	result.settingsBundleName = settingsBundleName
	result.settings = config.SettingsBundle(settingsBundleName)
	if result.settings == nil {
		return result, fmt.Errorf("settings bundle %q not found in model.NewLinksAPIHandlerParams", settingsBundleName)
	}
	result.linksCache = lc.MakeNullCache(result.settings.Harvester, result.settings.Harvester, result.settings.Harvester)
	result.progressReporter = result.settings.Observe.ProgressReporter()

	return result, nil
}

func (p defaultLinksAPIHandlerParams) Source() *APISource {
	return p.source
}

func (p defaultLinksAPIHandlerParams) Settings() *SettingsBundle {
	return p.settings
}

func (p defaultLinksAPIHandlerParams) LinksCache() lc.Cache {
	return p.linksCache
}

func (p defaultLinksAPIHandlerParams) ProgressReporter() observe.ProgressReporter {
	return p.progressReporter
}
