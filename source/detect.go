package source

import (
	"fmt"
	"regexp"

	"github.com/lectio/graph/model"
)

// UnknownSourceName is the value used for model.ContentSource.Name when the source is unknown
const UnknownSourceName = "Unidentified"

// DetectFromURLText detects the ContentSource given a URL string
func DetectFromURLText(source model.URLText) (model.ContentSource, LinksAPIHandlerFunc, error) {
	result := model.BookmarksAPISource{}
	dropmark := regexp.MustCompile(`^https\://(.*).dropmark.com/([0-9]+).json$`)

	switch {
	case dropmark.MatchString(string(source)):
		result.Name = "Dropmark"
		result.APIEndpoint = source
		return &result, DropmarkLinks, nil
	default:
		result.Name = UnknownSourceName
		result.APIEndpoint = source
		return &result, nil, fmt.Errorf("unable to detect %q as a valid ContentSource in source.DetectFromURLText", source)
	}
}

// DetectAPIFromURLText detects the APISource given a URL string
func DetectAPIFromURLText(sourceURL model.URLText) (model.APISource, LinksAPIHandlerFunc, error) {
	contentSource, handler, srcErr := DetectFromURLText(sourceURL)
	if srcErr != nil {
		return nil, nil, srcErr
	}

	apiSource, ok := contentSource.(model.APISource)
	if !ok {
		return nil, nil, fmt.Errorf("sourceURL %q did not resolve to a *model.APISource in source.DetectAPIFromURLText(): %+v", sourceURL, apiSource)
	}

	return apiSource, handler, nil
}
