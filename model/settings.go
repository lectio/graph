package model

import (
	graphql "github.com/99designs/gqlgen/graphql"
	"io"
	"net/http"

	"github.com/lectio/graph/observe"
	"github.com/lectio/resource"
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

// Configuration is the definition of all available settings bundles
type Configuration struct {
	defaultHTTPClient *http.Client
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
	c.defaultHTTPClient = &http.Client{Timeout: resource.HTTPTimeout}
	c.defaultStore = SettingsStore{Name: DefaultSettingsStoreName}
	c.linksStore = make(map[SettingsStoreName]*LinkLifecyleSettings)
	c.contentStore = make(map[SettingsStoreName]*ContentSettings)
	c.httpClientStore = make(map[SettingsStoreName]*HTTPClientSettings)
	c.repositoriesStore = make(map[SettingsStoreName]*Repositories)
	c.markdownGenStore = make(map[SettingsStoreName]*MarkdownGeneratorSettings)
}

// HTTPUserAgent returns the default HTTP user agent
func (c Configuration) HTTPUserAgent() string {
	return "github.com/lectio"
}

// HTTPClient returns the default HTTP client
func (c Configuration) HTTPClient() *http.Client {
	return c.defaultHTTPClient
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

	linkLC.TraverseLinks = true
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

	mdgSettings := new(MarkdownGeneratorSettings)
	mdgSettings.Store = c.defaultStore
	c.markdownGenStore[mdgSettings.Store.Name] = mdgSettings
	mdgSettings.CancelOnWriteErrors = 10
	mdgSettings.ContentPath = "content/post"
	mdgSettings.ImagesPath = "static/img/content/post"
	mdgSettings.ImagesURLRel = "/img/content/post"
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
