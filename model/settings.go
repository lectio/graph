package model

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	// "github.com/spf13/viper"
)

const (
	defaultSettingsBundleName SettingsBundleName = "DEFAULT"
)

// FollowRedirectsInDestinationHTMLContent defines whether we follow redirect rules in HTML <meta> refresh tags
func (lhs LinkHarvesterSettings) FollowRedirectsInDestinationHTMLContent(url *url.URL) bool {
	return lhs.FollowRedirectsInLinkDestinationHTMLContent
}

// ParseMetaDataInDestinationHTMLContent should be true if OpenGraph, TwitterCard, or other HTML meta data is required
func (lhs LinkHarvesterSettings) ParseMetaDataInDestinationHTMLContent(url *url.URL) bool {
	return lhs.ParseMetaDataInLinkDestinationHTMLContent
}

// DownloadAttachmentsFromDestination defines whether we download link attachments
func (lhs LinkHarvesterSettings) DownloadAttachmentsFromDestination(url *url.URL) (bool, string) {
	var destPath string // if this is blank, attachments are placed in temp directory
	return lhs.DownloadLinkDestinationAttachments, destPath
}

// IgnoreLink returns true (and a reason) if the given url should be ignored by the harvester
func (lhs LinkHarvesterSettings) IgnoreLink(url *url.URL) (bool, string) {
	URLtext := url.String()
	for _, regEx := range lhs.IgnoreURLsRegExprs {
		if regEx.MatchString(URLtext) {
			return true, fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
		}
	}
	return false, ""
}

// CleanLinkParams returns true if the given url's query string param should be "cleaned" by the harvester
func (lhs LinkHarvesterSettings) CleanLinkParams(url *url.URL) bool {
	// we try to clean all URLs, not specific ones
	return true
}

// RemoveQueryParamFromLinkURL returns true (and a reason) if the given url's specific query string param should be "cleaned" by the harvester
func (lhs LinkHarvesterSettings) RemoveQueryParamFromLinkURL(url *url.URL, paramName string) (bool, string) {
	for _, regEx := range lhs.RemoveParamsFromURLsRegEx {
		if regEx.MatchString(paramName) {
			return true, fmt.Sprintf("Matched cleaner rule %q: %q", regEx.String(), url.String())
		}
	}

	return false, ""
}

// PrimaryKeyForURL returns a globally unique key for the given URL (satisfies link.Keys interface)
func (lhs LinkHarvesterSettings) PrimaryKeyForURL(url *url.URL) string {
	if url != nil {
		return lhs.PrimaryKeyForURLText(url.String())
	}
	return "url_is_nil_in_PrimaryKeyForURL"
}

// PrimaryKeyForURLText returns a globally unique key for the given URL text (satisfies link.Keys interface)
func (lhs LinkHarvesterSettings) PrimaryKeyForURLText(urlText string) string {
	// TODO: consider adding a key cache since sha1 is compute intensive
	h := sha1.New()
	h.Write([]byte(urlText))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
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

	result.Observe.ProgressReporterType = ProgressReporterTypeCommandLineProgressBar

	result.Harvester.FollowRedirectsInLinkDestinationHTMLContent = true
	result.Harvester.SkipURLHumanMessageFormat.UnmarshalGQL("Skipping %[1]q: %[2]s") // 1 is the URL, 2 is the human readable reason
	result.Harvester.ParseMetaDataInLinkDestinationHTMLContent = true
	result.Harvester.DownloadLinkDestinationAttachments = false
	result.Harvester.InvalidLinksPolicy = InvalidLinksPolicySkipWithError
	result.Harvester.DuplicateLinksPolicy = DuplicatesRetentionPolicyRetainFirstSkipRemaining

	result.HTTPClient.UserAgent = "github.com/lectio/graph"
	result.HTTPClient.Timeout.UnmarshalGQL("90s")

	result.Content.Title.PipedSuffixPolicy = ContentTitleSuffixPolicyRemove
	result.Content.Title.HyphenatedSuffixPolicy = ContentTitleSuffixPolicyWarnIfDetected
	result.Content.Summary.Policy = ContentSummaryPolicyUseFirstSentenceOfContentBodyIfEmpty
	result.Content.Body.AllowFrontmatter = true
	result.Content.Body.FrontMatterPropertyNamePrefix = "body."

	return result
}

/*
func NewViperConfiguration(h *ServiceHandler, provider ConfigPathProvider, configName SettingsBundleName, parent opentracing.Span) *Settings {
	span := h.observatory.StartChildTrace("resolvers.NewViperConfiguration", parent)
	defer span.Finish()
``
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
