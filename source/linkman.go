package source

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/lectio/graph/model"
	"github.com/lectio/resource"

	"github.com/lectio/link"
)

// simpleLink is a link.Link instance that does not do any traversals or other magic (satisfies link.Link interface)
type simpleLink string

func (l simpleLink) OriginalURL() string {
	return string(l)
}

func (l simpleLink) FinalURL() (*url.URL, error) {
	return url.Parse(string(l))
}

// LinksManager wraps LinkLifecyleSettings and implements a number of interfaces
type LinksManager struct {
	Config       *model.Configuration
	LinkSettings *model.LinkLifecyleSettings
	Client       *http.Client
}

// HTTPClient defines the HTTP client for the link destination to use
// This method satisfies resource.Policy interface
func (lm LinksManager) HTTPClient() *http.Client {
	return lm.Client
}

// PrepareRequest satisfies resource.Policy interface
func (lm LinksManager) PrepareRequest(client *http.Client, req *http.Request) {
	req.Header.Set("User-Agent", "github.com/lectio/source.LinksManager")
}

// DetectRedirectsInHTMLContent defines whether we detect redirect rules in HTML <meta> refresh tags
// This method satisfies resource.Policy interface
func (lm LinksManager) DetectRedirectsInHTMLContent(*url.URL) bool {
	return lm.LinkSettings.FollowRedirectsInLinkDestinationHTMLContent
}

// FollowRedirectsInHTMLContent defines whether we follow redirect rules in HTML <meta> refresh tags
func (lm LinksManager) FollowRedirectsInHTMLContent(url *url.URL) bool {
	return lm.LinkSettings.FollowRedirectsInLinkDestinationHTMLContent
}

// ParseMetaDataInHTMLContent defines whether we want to parse HTML meta data
// This method satisfies resource.Policy interface
func (lm LinksManager) ParseMetaDataInHTMLContent(*url.URL) bool {
	return lm.LinkSettings.ParseMetaDataInLinkDestinationHTMLContent
}

// DownloadContent satisfies Policy method
func (lm LinksManager) DownloadContent(url *url.URL, resp *http.Response, typ resource.Type) (bool, resource.Attachment, []resource.Issue) {
	if !lm.LinkSettings.DownloadLinkDestinationAttachments {
		return false, nil, nil
	}
	return resource.DownloadFile(lm, url, resp, typ)
}

// CreateFile satisfies FileAttachmentPolicy method
func (lm LinksManager) CreateFile(url *url.URL, t resource.Type) (*os.File, resource.Issue) {
	return nil, resource.NewIssue(url.String(), "NOT_IMPLEMENTED_YET", "CreateFile not implemented in graph.LinkLifecyleSettings", true)
}

// AutoAssignExtension satisfies FileAttachmentPolicy method
func (lm LinksManager) AutoAssignExtension(url *url.URL, t resource.Type) bool {
	return true
}

// IgnoreLink returns true (and a reason) if the given url should be ignored by the harvester
func (lm LinksManager) IgnoreLink(url *url.URL) (bool, string) {
	URLtext := url.String()
	for _, regEx := range lm.LinkSettings.IgnoreURLsRegExprs {
		if regEx.MatchString(URLtext) {
			return true, fmt.Sprintf("Matched Ignore Rule `%s`", regEx.String())
		}
	}
	return false, ""
}

// CleanLinkParams returns true if the given url's query string param should be "cleaned" by the harvester
func (lm LinksManager) CleanLinkParams(url *url.URL) bool {
	// we try to clean all URLs, not specific ones
	return true
}

// RemoveQueryParamFromLinkURL returns true (and a reason) if the given url's specific query string param should be "cleaned" by the harvester
func (lm LinksManager) RemoveQueryParamFromLinkURL(url *url.URL, paramName string) (bool, string) {
	for _, regEx := range lm.LinkSettings.RemoveParamsFromURLsRegEx {
		if regEx.MatchString(paramName) {
			return true, fmt.Sprintf("Matched cleaner rule %q: %q", regEx.String(), url.String())
		}
	}

	return false, ""
}

// PrimaryKeyForURL returns a globally unique key for the given URL (satisfies link.Keys interface)
func (lm LinksManager) PrimaryKeyForURL(url *url.URL) string {
	if url != nil {
		return lm.PrimaryKeyForURLText(url.String())
	}
	return "url_is_nil_in_PrimaryKeyForURL"
}

// PrimaryKeyForURLText returns a globally unique key for the given URL text (satisfies link.Keys interface)
func (lm LinksManager) PrimaryKeyForURLText(urlText string) string {
	// TODO: consider adding a key cache since sha1 is compute intensive
	h := sha1.New()
	h.Write([]byte(urlText))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// HarvestLink satisfies the link.Lifecyle interface and creates a new Link from a URL string
func (lm LinksManager) HarvestLink(urlText string) (link.Link, link.Issue) {
	if lm.LinkSettings.TraverseLinks {
		return link.TraverseLink(urlText, lm, lm, lm), nil
	}
	sl := simpleLink(urlText)
	return sl, nil
}
