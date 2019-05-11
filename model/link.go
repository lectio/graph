package model

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

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
