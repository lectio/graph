package model

import (
	"fmt"
	"net/url"
	// "github.com/spf13/viper"
)

const (
	defaultSettingsBundleName SettingsBundleName = "DEFAULT"
)

// GlobalConfiguration is the package global settings bundles instance
var GlobalConfiguration, _ = MakeConfiguration()

// IgnoreResource implements the github.com/lectio/link.CleanResourceParamsRule interface
func (lfsb LinkHarvesterSettings) IgnoreResource(url *url.URL) (bool, string) {
	URLtext := url.String()
	for _, regEx := range lfsb.IgnoreURLsRegExprs {
		if regEx.MatchString(URLtext) {
			return true, fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
		}
	}
	return false, ""
}

// CleanResourceParams implements the github.com/lectio/link.CleanResourceParamsRule interface
func (lfsb LinkHarvesterSettings) CleanResourceParams(url *url.URL) bool {
	// we try to clean all URLs, not specific ones
	return true
}

// RemoveQueryParamFromResourceURL implements the github.com/lectio/link.CleanResourceParamsRule interface
func (lfsb LinkHarvesterSettings) RemoveQueryParamFromResourceURL(paramName string) (bool, string) {
	for _, regEx := range lfsb.RemoveParamsFromURLsRegEx {
		if regEx.MatchString(paramName) {
			return true, fmt.Sprintf("Matched cleaner rule `%s`", regEx.String())
		}
	}
	return false, ""
}

// Configuration is the definition of all available settings bundles
type Configuration struct {
	bundles       map[SettingsBundleName]*SettingsBundle
	defaultBundle *SettingsBundle
}

// MakeConfiguration creates a new SettingsBundle instance with default options
func MakeConfiguration() (*Configuration, error) {
	result := new(Configuration)
	result.defaultBundle = result.createDefaultBundle()
	result.bundles = make(map[SettingsBundleName]*SettingsBundle)
	result.bundles[result.defaultBundle.Name] = result.defaultBundle
	return result, nil
}

// SettingsBundle returns a named settings bundle
func (c Configuration) SettingsBundle(name SettingsBundleName) *SettingsBundle {
	return c.bundles[name]
}

// DefaultBundle returns the default settings bundle
func (c Configuration) DefaultBundle() *SettingsBundle {
	return c.defaultBundle
}

// Close frees up any resources allocated by the settings bundles instance
func (c *Configuration) Close() {
	// nothing to close
}

func (c *Configuration) createDefaultBundle() *SettingsBundle {
	result := new(SettingsBundle)
	result.Name = defaultSettingsBundleName

	re, err := MakeRegularExpression(`^https://twitter.com/(.*?)/status/(.*)$`)
	if err == nil {
		result.Harvester.IgnoreURLsRegExprs = append(result.Harvester.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`https://t.co`)
	if err == nil {
		result.Harvester.IgnoreURLsRegExprs = append(result.Harvester.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`^utm_`)
	if err == nil {
		result.Harvester.RemoveParamsFromURLsRegEx = append(result.Harvester.RemoveParamsFromURLsRegEx, re)
	} else {
		panic(err)
	}

	result.Harvester.FollowHTMLRedirects = true
	result.Harvester.DuplicateLinkRetentionType = DuplicateRetentionTypeRetainAllButWarnOnDuplicate
	result.Harvester.SkipURLHumanMessageFormat.UnmarshalGQL("Skipping %[1]q: %[2]s") // 1 is the URL, 2 is the human readable reason
	result.Harvester.InspectLinkDestinations = true
	result.Harvester.DownloadLinkAttachments = false

	result.HTTPClient.UserAgent = "github.com/lectio/graph"
	result.HTTPClient.Timeout.UnmarshalGQL("90s")

	result.Content.Title.RemovePipedSuffix = true
	result.Content.Title.WarnAboutHyphenatedSuffix = true
	result.Content.Summary.UseFirstSentenceOfBodyIfEmpty = true
	result.Content.Body.AllowFrontmatter = true
	result.Content.Body.FrontMatterPropertyNamePrefix = "body."

	return result
}

/*
func NewViperConfiguration(h *ServiceHandler, provider ConfigPathProvider, configName SettingsBundleName, parent opentracing.Span) *Settings {
	span := h.observatory.StartChildTrace("resolvers.NewViperConfiguration", parent)
	defer span.Finish()

	result := new(Configuration)
	v := viper.New()

	v.SetEnvPrefix("LECTIO_GRAPH_CONF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName(string(configName))
	for _, path := range provider(string(configName)) {
		v.AddConfigPath(path)
	}
	err := v.ReadInConfig()

	if err != nil {
		opentrext.Error.Set(span, true)
		span.LogFields(log.Error(err))
	} else {
		span.LogFields(log.String("Read configuration from file %s", v.ConfigFileUsed()))
		err = v.Unmarshal(&result.settings)
		if err != nil {
			opentrext.Error.Set(span, true)
			span.LogFields(log.Error(err))
		}
	}

	result.ConfigureContentHarvester(h, parent)
	return result
}
*/

// // ConfigureContentHarvester uses the config parameters in Configuration().Harvest to setup the content harvester
// func (c *Configuration) ConfigureContentHarvester(h *ServiceHandler, parent opentracing.Span) {
// 	span := h.observatory.StartChildTrace("resolvers.ConfigureContentHarvester", parent)
// 	defer span.Finish()

// 	c.store = persistence.NewDatastore(h.observatory, &c.settings.Storage, span)
// 	c.ignoreURLsRegEx.AddMultiple(c.settings, c.settings.Harvest.IgnoreURLsRegExprs)
// 	c.removeParamsFromURLsRegEx.AddMultiple(c.settings, c.settings.Harvest.RemoveParamsFromURLsRegEx)
// 	c.contentHarvester = harvester.MakeContentHarvester(h.observatory, c.ignoreURLsRegEx, c.removeParamsFromURLsRegEx, c.settings.Harvest.FollowHTMLRedirects)
// }
