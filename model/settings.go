package model

import (
	"fmt"
	graphql "github.com/99designs/gqlgen/graphql"
	"io"
	"net/http"
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
	defaultStore             SettingsStore
	linksSettingsStore       map[SettingsStoreName]*LinkLifecyleSettings
	contentSettingsStore     map[SettingsStoreName]*ContentSettings
	httpClientSettingsStore  map[SettingsStoreName]*HTTPClientSettings
	httpClients              map[SettingsStoreName]*http.Client
	repositoriesStore        map[SettingsStoreName]*Repositories
	markdownGenStore         map[SettingsStoreName]*MarkdownGeneratorSettings
	observationSettingsStore map[SettingsStoreName]*ObservationSettings
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
	c.linksSettingsStore = make(map[SettingsStoreName]*LinkLifecyleSettings)
	c.contentSettingsStore = make(map[SettingsStoreName]*ContentSettings)
	c.httpClientSettingsStore = make(map[SettingsStoreName]*HTTPClientSettings)
	c.httpClients = make(map[SettingsStoreName]*http.Client)
	c.repositoriesStore = make(map[SettingsStoreName]*Repositories)
	c.markdownGenStore = make(map[SettingsStoreName]*MarkdownGeneratorSettings)
	c.observationSettingsStore = make(map[SettingsStoreName]*ObservationSettings)
}

// HTTPClient returns an HTTP client associated with the given path
func (c Configuration) HTTPClient(path SettingsPath) *http.Client {
	return c.httpClients[SettingsStoreName(path)]
}

// LinkLifecyleSettings returns the first LinkLifecyleSettings found in path, or the default (should never be nil)
func (c Configuration) LinkLifecyleSettings(path SettingsPath) *LinkLifecyleSettings {
	return c.linksSettingsStore[SettingsStoreName(path)]
}

// ContentSettings returns the first ContentSettings found in path, or the default (should never be nil)
func (c Configuration) ContentSettings(path SettingsPath) *ContentSettings {
	return c.contentSettingsStore[SettingsStoreName(path)]
}

// HTTPClientSettings returns the first HTTPClientSettings found in path, or the default (should never be nil)
func (c Configuration) HTTPClientSettings(path SettingsPath) *HTTPClientSettings {
	return c.httpClientSettingsStore[SettingsStoreName(path)]
}

// Repositories returns the first Repositories found in path, or the default (should never be nil)
func (c Configuration) Repositories(path SettingsPath) *Repositories {
	return c.repositoriesStore[SettingsStoreName(path)]
}

// MarkdownGeneratorSettings returns the first MarkdownGeneratorSettings found in path, or the default (should never be nil)
func (c Configuration) MarkdownGeneratorSettings(path SettingsPath) *MarkdownGeneratorSettings {
	return c.markdownGenStore[SettingsStoreName(path)]
}

// ObservationSettings returns the first ObservationSettings found in path, or the default (should never be nil)
func (c Configuration) ObservationSettings(path SettingsPath) *ObservationSettings {
	return c.observationSettingsStore[SettingsStoreName(path)]
}

// Close frees up any resources allocated by the settings bundles instance
func (c *Configuration) Close() {
	// nothing to close
}

func (c *Configuration) createDefaults() {
	linkLC := new(LinkLifecyleSettings)
	linkLC.Store = c.defaultStore
	c.linksSettingsStore[linkLC.Store.Name] = linkLC

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
	c.httpClientSettingsStore[httpClientSettings.Store.Name] = httpClientSettings
	httpClientSettings.UserAgent = "github.com/lectio/graph"
	httpClientSettings.Timeout.UnmarshalGQL("90s")

	// You can use local HTTP caching but it does not seem to any faster than just running without cache for Link traversal and image downloads.
	// But, perhaps it might be faster for social scoring? Need to try it out.
	//httpClientSettings.Cache = &HTTPDiskCache{Name: "HTTPDiskCache:support/tmp/httpcache", BasePath: "support/tmp/httpcache", CreateBasePath: true}
	//httpClientSettings.Cache = &HTTPMemoryCache{Name: "HTTPMemoryCache"}

	httpClient, hcErr := httpClientSettings.NewHTTPClient()
	if hcErr != nil {
		// Even if error is encountered, the NewHTTPClient() method should return a valid http.Client
		fmt.Printf("Unable to create HTTP Client: %s, using default instead.\n" + hcErr.Error())
	}
	c.httpClients[httpClientSettings.Store.Name] = httpClient

	contentSettings := new(ContentSettings)
	contentSettings.Store = c.defaultStore
	c.contentSettingsStore[contentSettings.Store.Name] = contentSettings
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

	obsSettings := new(ObservationSettings)
	obsSettings.Store = c.defaultStore
	c.observationSettingsStore[mdgSettings.Store.Name] = obsSettings
	obsSettings.ProgressReporterType = ProgressReporterTypeProgressBar
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
	for _, v := range c.linksSettingsStore {
		result = append(result, v)
	}
	for _, v := range c.contentSettingsStore {
		result = append(result, v)
	}
	for _, v := range c.httpClientSettingsStore {
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
