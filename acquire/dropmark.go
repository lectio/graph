package acquire

import (
	"fmt"
	"time"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// DropmarkLinks returns a collection of harvested links from a Dropmark API
func DropmarkLinks(params model.LinksAPIHandlerParams) (*model.Bookmarks, error) {
	source, ok := params.Source().(*model.BookmarksAPISource)
	if !ok {
		return nil, fmt.Errorf("Source is %+v, acquire.DropmarkLinks requires a model.BookmarksAPISource", params.Source())
	}

	pr := params.ProgressReporter()
	settings := params.Settings()

	dc, dcErr := dropmark.GetCollection(string(source.APIEndpoint), pr, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	dropColl := model.Bookmarks{}
	dropColl.Source = *source
	dropColl.Activities = model.Activities{}

	work := func(ch chan<- int, index int, item *dropmark.Item) {
		bookmark := model.Bookmark{
			ID:         settings.Links.PrimaryKeyForURLText(item.Link),
			Link:       model.BookmarkLink{OriginalURLText: model.URLText(item.Link)},
			Title:      model.ContentTitleText(item.Name),
			Summary:    model.ContentSummaryText(item.Description),
			Body:       model.ContentBodyText(item.Content),
			Properties: model.MakeProperties()}

		bookmark.Title.Edit(&bookmark, &settings.Content.Title)
		bookmark.Summary.Edit(&bookmark, &settings.Content.Summary)
		bookmark.Body.Edit(&bookmark, &settings.Content.Body)

		if bookmark.Link.OriginalURLText.IsEmpty() {
			dropColl.Activities.AddWarning(string(source.APIEndpoint), "DLWARN-0101-LINKEMPTY", fmt.Sprintf("Dropmark link %d is invalid: URL is %q", index, item.Link))
			ch <- index
			return
		}

		link, linkErr := bookmark.Link.OriginalURLText.Link(params.Settings())
		if linkErr != nil || link == nil {
			dropColl.Activities.AddError(string(source.APIEndpoint), "DLERR-0101-LINKERR", fmt.Sprintf("Dropmark link %d (%q) is either nil or invalid: %v", index, item.Link, linkErr))
			ch <- index
			return
		}

		finalURL, finalErr := link.FinalURL()
		if finalErr != nil {
			dropColl.Activities.AddError(string(source.APIEndpoint), "DLERR-0100-FINALURL", finalErr.Error())
			ch <- index
			return
		}

		ignore, ignoreReason := link.Ignore()
		if ignore {
			dropColl.Activities.AddWarning(string(source.APIEndpoint), "DLWARN-0100-IGNORE", ignoreReason)
			ch <- index
			return
		}

		bookmark.ID = link.PrimaryKey(settings.Links)
		bookmark.Link.IsValid = true
		bookmark.Link.FinalURL = model.MakeURL(finalURL)
		dropColl.Content = append(dropColl.Content, bookmark)
		ch <- index
	}

	if pr != nil && pr.IsProgressReportingRequested() {
		pr.StartReportableActivity(len(dc.Items))
	}
	ch := make(chan int)
	for index, item := range dc.Items {
		go work(ch, index, item)
	}
	for range dc.Items {
		_ = <-ch
		if pr != nil && pr.IsProgressReportingRequested() {
			pr.IncrementReportableActivityProgress()
		}
	}
	if pr != nil && pr.IsProgressReportingRequested() {
		pr.CompleteReportableActivityProgress(fmt.Sprintf("Completed creating %d %s Links from %q", len(dropColl.Content), source.Name, source.APIEndpoint))
	}
	return &dropColl, nil
}
