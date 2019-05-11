package model

import (
	"crypto/sha1"
	"fmt"
	graphql "github.com/99designs/gqlgen/graphql"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/lectio/graph/observe"
	"github.com/lectio/resource"
	"github.com/lectio/score"

	"github.com/lectio/link"
)

const (
	// DefaultSettingsStoreName is the built-in persistent settings store
	DefaultSettingsStoreName SettingsStoreName = "DEFAULT"

	// DefaultSettingsPath is always available and used when a custom settings path is not supplied
	DefaultSettingsPath SettingsPath = "DEFAULT"
)

// SettingsPath is a UNIX path-like string delimited using :: for instructing which settings should be used
type SettingsPath string

// SettingsStoreName is the name of a single settings store (it could be a URL or local filename)
type SettingsStoreName string

// MarshalGQL emits GraphQL
func (t SettingsPath) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

// UnmarshalGQL converts GraphQL to Go
func (t *SettingsPath) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = SettingsPath(str)
	}
	return err
}

// MarshalGQL emits GraphQL
func (t SettingsStoreName) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

// UnmarshalGQL converts GraphQL to Go
func (t *SettingsStoreName) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = SettingsStoreName(str)
	}
	return err
}

// simpleLink is a link.Link instance that does not do any traversals or other magic (satisfies link.Link interface)
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
func (lhs LinkLifecyleSettings) HarvestLink(urlText string) (link.Link, link.Issue) {
	if lhs.TraverseLinks {
		return link.TraverseLink(urlText, lhs, lhs, lhs), nil
	} else {
		sl := simpleLink(urlText)
		return sl, nil
	}
}

// ScoreLink socially scores the given URL
func (lhs LinkLifecyleSettings) ScoreLink(url *url.URL) (score.LinkScores, score.Issue) {
	vault, vaultErr := MakeSecretsVault("env://LECTIO_VAULTPP_DEFAULT")
	if vaultErr != nil {
		panic(vaultErr)
	}
	if lhs.ScoreLinks.Score {
		scores, err := score.GetSharedCountLinkScoresForURL(vault, url, lhs, lhs.ScoreLinks.Simulate)
		if err != nil {
			return nil, score.NewIssue("SharedCount.com", "API error", err.Error(), true)
		}
		return scores, nil
	}

	return nil, nil
}

// Configuration is the definition of all available settings bundles
type Configuration struct {
	defaultStore      SettingsStore
	linksStore        map[SettingsStoreName]*LinkLifecyleSettings
	contentStore      map[SettingsStoreName]*ContentSettings
	httpClientStore   map[SettingsStoreName]*HTTPClientSettings
	repositoriesStore map[SettingsStoreName]*Repositories
	markdownGenStore  map[SettingsStoreName]*MarkdownGeneratorSettings
}

// MakeConfiguration creates a new SettingsBundle instance with default options
func MakeConfiguration() (*Configuration, error) {
	result := new(Configuration)
	result.init()
	result.createDefaults()
	return result, nil
}

func (c *Configuration) init() {
	c.defaultStore = SettingsStore{Name: DefaultSettingsStoreName}
	c.linksStore = make(map[SettingsStoreName]*LinkLifecyleSettings)
	c.contentStore = make(map[SettingsStoreName]*ContentSettings)
	c.httpClientStore = make(map[SettingsStoreName]*HTTPClientSettings)
	c.repositoriesStore = make(map[SettingsStoreName]*Repositories)
	c.markdownGenStore = make(map[SettingsStoreName]*MarkdownGeneratorSettings)
}

// ProgressReporter returns the observation strategy
func (c Configuration) ProgressReporter() observe.ProgressReporter {
	return observe.DefaultCommandLineProgressReporter
}

// LinkLifecyleSettings returns the first LinkLifecyleSettings found in path, or the default (should never be nil)
func (c Configuration) LinkLifecyleSettings(path SettingsPath) *LinkLifecyleSettings {
	return c.linksStore[SettingsStoreName(path)]
}

// ContentSettings returns the first ContentSettings found in path, or the default (should never be nil)
func (c Configuration) ContentSettings(path SettingsPath) *ContentSettings {
	return c.contentStore[SettingsStoreName(path)]
}

// HTTPClientSettings returns the first HTTPClientSettings found in path, or the default (should never be nil)
func (c Configuration) HTTPClientSettings(path SettingsPath) *HTTPClientSettings {
	return c.httpClientStore[SettingsStoreName(path)]
}

// Repositories returns the first Repositories found in path, or the default (should never be nil)
func (c Configuration) Repositories(path SettingsPath) *Repositories {
	return c.repositoriesStore[SettingsStoreName(path)]
}

// MarkdownGeneratorSettings returns the first MarkdownGeneratorSettings found in path, or the default (should never be nil)
func (c Configuration) MarkdownGeneratorSettings(path SettingsPath) *MarkdownGeneratorSettings {
	return c.markdownGenStore[SettingsStoreName(path)]
}

// Close frees up any resources allocated by the settings bundles instance
func (c *Configuration) Close() {
	// nothing to close
}

func (c *Configuration) createDefaults() {
	linkLC := new(LinkLifecyleSettings)
	linkLC.Store = c.defaultStore
	c.linksStore[linkLC.Store.Name] = linkLC

	re, err := MakeRegularExpression(`^https://twitter.com/(.*?)/status/(.*)$`)
	if err == nil {
		linkLC.IgnoreURLsRegExprs = append(linkLC.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`https://t.co`)
	if err == nil {
		linkLC.IgnoreURLsRegExprs = append(linkLC.IgnoreURLsRegExprs, re)
	} else {
		panic(err)
	}

	re, err = MakeRegularExpression(`^utm_`)
	if err == nil {
		linkLC.RemoveParamsFromURLsRegEx = append(linkLC.RemoveParamsFromURLsRegEx, re)
	} else {
		panic(err)
	}

	linkLC.TraverseLinks = false
	linkLC.ScoreLinks.Score = true
	linkLC.ScoreLinks.Simulate = true
	linkLC.FollowRedirectsInLinkDestinationHTMLContent = true
	linkLC.ParseMetaDataInLinkDestinationHTMLContent = true
	linkLC.DownloadLinkDestinationAttachments = false

	repositories := new(Repositories)
	repositories.Store = c.defaultStore
	c.repositoriesStore[repositories.Store.Name] = repositories
	repositories.All = append(repositories.All, TempFileRepository{
		Name:   "TEMP",
		URL:    "file:///tmp",
		Prefix: "lectio_tmp"})

	httpClientSettings := new(HTTPClientSettings)
	httpClientSettings.Store = c.defaultStore
	c.httpClientStore[httpClientSettings.Store.Name] = httpClientSettings
	httpClientSettings.UserAgent = "github.com/lectio/graph"
	httpClientSettings.Timeout.UnmarshalGQL("90s")

	contentSettings := new(ContentSettings)
	contentSettings.Store = c.defaultStore
	c.contentStore[contentSettings.Store.Name] = contentSettings
	contentSettings.Title.PipedSuffixPolicy = ContentTitleSuffixPolicyRemove
	contentSettings.Title.HyphenatedSuffixPolicy = ContentTitleSuffixPolicyWarnIfDetected
	contentSettings.Summary.Policy = ContentSummaryPolicyUseFirstSentenceOfContentBodyIfEmpty
	contentSettings.Body.AllowFrontmatter = true
	contentSettings.Body.FrontMatterPropertyNamePrefix = "body."

	hugoSettings := new(MarkdownGeneratorSettings)
	hugoSettings.Store = c.defaultStore
	c.markdownGenStore[hugoSettings.Store.Name] = hugoSettings
	hugoSettings.CancelOnWriteErrors = 10
	hugoSettings.ContentPath = "content/post"
	hugoSettings.ImagesPath = "static/img/content/post"
	hugoSettings.ImagesURLRel = "/img/content/post"
}

// Vault returns the default secrets valut
func (c Configuration) Vault() *SecretsVault {
	vault, vaultErr := MakeSecretsVault("env://LECTIO_VAULTPP_DEFAULT")
	if vaultErr != nil {
		panic(vaultErr)
	}
	return vault
}

// AllSettings returns all known settings
func (c Configuration) AllSettings() ([]PersistentSettings, error) {
	var result []PersistentSettings
	for _, v := range c.linksStore {
		result = append(result, v)
	}
	for _, v := range c.contentStore {
		result = append(result, v)
	}
	for _, v := range c.httpClientStore {
		result = append(result, v)
	}
	for _, v := range c.repositoriesStore {
		result = append(result, v)
	}
	for _, v := range c.markdownGenStore {
		result = append(result, v)
	}
	return result, nil
}
