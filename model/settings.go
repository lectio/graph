package model

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/lectio/resource"
	"github.com/lectio/score"

	"github.com/lectio/link"
	// "github.com/spf13/viper"
)

const (
	// DefaultSettingsBundleName is always available and used when a custom settings bundle is not required
	DefaultSettingsBundleName SettingsBundleName = "DEFAULT"
)

// simpleLink is a link.Link instance that does not do any traversals or other magic
type simpleLink string

func (l simpleLink) OriginalURL() string {
	return string(l)
}

func (l simpleLink) FinalURL() (*url.URL, error) {
	return url.Parse(string(l))
}

// HTTPUserAgent defines the HTTP GET user agent
// This method satisfies resource.Policy interface
func (lhs LinkLifecyleSettings) HTTPUserAgent() string {
	return "github.com/lectio/graph/model"
}

// HTTPTimeout defines the HTTP GET timeout duration
// This method satisfies resource.Policy interface
func (lhs LinkLifecyleSettings) HTTPTimeout() time.Duration {
	return resource.HTTPTimeout
}

// DetectRedirectsInHTMLContent defines whether we detect redirect rules in HTML <meta> refresh tags
// This method satisfies resource.Policy interface
func (lhs LinkLifecyleSettings) DetectRedirectsInHTMLContent(*url.URL) bool {
	return lhs.FollowRedirectsInLinkDestinationHTMLContent
}

// FollowRedirectsInHTMLContent defines whether we follow redirect rules in HTML <meta> refresh tags
func (lhs LinkLifecyleSettings) FollowRedirectsInHTMLContent(url *url.URL) bool {
	return lhs.FollowRedirectsInLinkDestinationHTMLContent
}

// ParseMetaDataInHTMLContent defines whether we want to parse HTML meta data
// This method satisfies resource.Policy interface
func (lhs LinkLifecyleSettings) ParseMetaDataInHTMLContent(*url.URL) bool {
	return lhs.ParseMetaDataInLinkDestinationHTMLContent
}

// DownloadContent satisfies Policy method
func (lhs LinkLifecyleSettings) DownloadContent(url *url.URL, resp *http.Response, typ resource.Type) (bool, resource.Attachment, []resource.Issue) {
	if !lhs.DownloadLinkDestinationAttachments {
		return false, nil, nil
	}
	return resource.DownloadFile(lhs, url, resp, typ)
}

// CreateFile satisfies FileAttachmentPolicy method
func (lhs LinkLifecyleSettings) CreateFile(url *url.URL, t resource.Type) (*os.File, resource.Issue) {
	return nil, resource.NewIssue(url.String(), "NOT_IMPLEMENTED_YET", "CreateFile not implemented in graph.LinkLifecyleSettings", true)
}

// AutoAssignExtension satisfies FileAttachmentPolicy method
func (lhs LinkLifecyleSettings) AutoAssignExtension(url *url.URL, t resource.Type) bool {
	return true
}

// IgnoreLink returns true (and a reason) if the given url should be ignored by the harvester
func (lhs LinkLifecyleSettings) IgnoreLink(url *url.URL) (bool, string) {
	URLtext := url.String()
	for _, regEx := range lhs.IgnoreURLsRegExprs {
		if regEx.MatchString(URLtext) {
			return true, fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
		}
	}
	return false, ""
}

// CleanLinkParams returns true if the given url's query string param should be "cleaned" by the harvester
func (lhs LinkLifecyleSettings) CleanLinkParams(url *url.URL) bool {
	// we try to clean all URLs, not specific ones
	return true
}

// RemoveQueryParamFromLinkURL returns true (and a reason) if the given url's specific query string param should be "cleaned" by the harvester
func (lhs LinkLifecyleSettings) RemoveQueryParamFromLinkURL(url *url.URL, paramName string) (bool, string) {
	for _, regEx := range lhs.RemoveParamsFromURLsRegEx {
		if regEx.MatchString(paramName) {
			return true, fmt.Sprintf("Matched cleaner rule %q: %q", regEx.String(), url.String())
		}
	}

	return false, ""
}

// PrimaryKeyForURL returns a globally unique key for the given URL (satisfies link.Keys interface)
func (lhs LinkLifecyleSettings) PrimaryKeyForURL(url *url.URL) string {
	if url != nil {
		return lhs.PrimaryKeyForURLText(url.String())
	}
	return "url_is_nil_in_PrimaryKeyForURL"
}

// PrimaryKeyForURLText returns a globally unique key for the given URL text (satisfies link.Keys interface)
func (lhs LinkLifecyleSettings) PrimaryKeyForURLText(urlText string) string {
	// TODO: consider adding a key cache since sha1 is compute intensive
	h := sha1.New()
	h.Write([]byte(urlText))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// HarvestLink satisfies the link.Lifecyle interface and creates a new Link from a URL string
func (sb SettingsBundle) HarvestLink(urlText string) (link.Link, link.Issue) {
	if sb.Links.TraverseLinks {
		return link.TraverseLink(urlText, sb.Links, sb.Links, sb.Links), nil
	} else {
		sl := simpleLink(urlText)
		return sl, nil
	}
}

// SharedCountAPIKey satisfies score.SharedCountCredentials interface for getting the SharedCount.com service credentials
func (sb SettingsBundle) SharedCountAPIKey() (string, bool, score.Issue) {
	apiKey, err := sb.Vault().vault.DecryptText("0d4af7674abbfa18d01510fc107318ace74175c5cae32b1e3dfb1ec37ee5ceb1c8253d880ba027ed3c8280883cef0152d447f068a21f0a793f83c552fd89703aeecd53d5")
	if err != nil {
		return "", false, score.NewIssue("SharedCount.com", score.SecretManagementError, err.Error(), true)
	}
	return apiKey, true, nil
}

// ScoreLink satisfies the score.Lifecyle interface and creates a new LinkScore from a URL string
func (sb SettingsBundle) ScoreLink(url *url.URL) (score.LinkScores, score.Issue) {
	if sb.Links.ScoreLinks.Score {
		scores, err := score.GetSharedCountLinkScoresForURL(sb, url, sb.Links, sb.Links.ScoreLinks.Simulate)
		if err != nil {
			return nil, score.NewIssue("SharedCount.com", "API error", err.Error(), true)
		}
		return scores, nil
	}

	return nil, nil
}

// Vault returns the default secrets valut
func (sb SettingsBundle) Vault() *SecretsVault {
	vault, vaultErr := MakeSecretsVault("env://LECTIO_VAULTPP_DEFAULT")
	if vaultErr != nil {
		panic(vaultErr)
	}
	return vault
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
		result.Links.IgnoreURLsRegExprs = append(result.Links.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`https://t.co`)
	if err == nil {
		result.Links.IgnoreURLsRegExprs = append(result.Links.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`^utm_`)
	if err == nil {
		result.Links.RemoveParamsFromURLsRegEx = append(result.Links.RemoveParamsFromURLsRegEx, re)
	} else {
		panic(err)
	}

	result.Observe.ProgressReporterType = ProgressReporterTypeCommandLineProgressBar

	vault, vaultErr := MakeSecretsVault("env://LECTIO_VAULTPP_DEFAULT")
	if vaultErr != nil {
		panic(vaultErr)
	}

	result.Repositories.All = append(result.Repositories.All, TempFileRepository{
		Name:   "TEMP",
		URL:    "file:///tmp",
		Prefix: "lectio_tmp"})

	result.Links.TraverseLinks = false
	result.Links.ScoreLinks.Score = true
	result.Links.ScoreLinks.Simulate = true
	result.Links.FollowRedirectsInLinkDestinationHTMLContent = true
	result.Links.ParseMetaDataInLinkDestinationHTMLContent = true
	result.Links.DownloadLinkDestinationAttachments = false

	result.HTTPClient.UserAgent = "github.com/lectio/graph"
	result.HTTPClient.Timeout.UnmarshalGQL("90s")

	result.Content.Title.PipedSuffixPolicy = ContentTitleSuffixPolicyRemove
	result.Content.Title.HyphenatedSuffixPolicy = ContentTitleSuffixPolicyWarnIfDetected
	result.Content.Summary.Policy = ContentSummaryPolicyUseFirstSentenceOfContentBodyIfEmpty
	result.Content.Body.AllowFrontmatter = true
	result.Content.Body.FrontMatterPropertyNamePrefix = "body."

	return result
}
