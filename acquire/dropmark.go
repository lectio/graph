package acquire

import (
	"fmt"
	"time"

	"github.com/lectio/dropmark"
	lp "github.com/lectio/link"

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

	dropColl := model.Bookmarks{}
	dropColl.Source = *source
	dropColl.Activities = model.Activities{}

	dc, issues := dropmark.GetCollection(string(source.APIEndpoint), pr, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if issues != nil {
		issues.HandleIssues(
			func(err dropmark.Issue) {
				dropColl.Activities.AddError(string(source.APIEndpoint), string(err.IssueCode()), err.Issue())
			},
			func(warning dropmark.Issue) {
				dropColl.Activities.AddWarning(string(source.APIEndpoint), string(warning.IssueCode()), warning.Issue())
			})
		return &dropColl, nil
	}

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

		issueContext := fmt.Sprintf("[%s] Dropmark link %d %q", source.APIEndpoint, index, item.Link)

		if bookmark.Link.OriginalURLText.IsEmpty() {
			dropColl.Activities.AddWarning(issueContext, "DLWARN-0101-LINKEMPTY", "Empty link")
			ch <- index
			return
		}

		link, linkErr := bookmark.Link.OriginalURLText.Link(params.Settings())
		if linkErr != nil || link == nil {
			dropColl.Activities.AddError(issueContext, "DLERR-0101-LINKERR", fmt.Sprintf("Unable to create link.Link: %v", linkErr))
			ch <- index
			return
		}

		if link.Issues() != nil {
			var exitOnErrors, exitOnWarnings int
			link.Issues().HandleIssues(
				func(err lp.Issue) {
					dropColl.Activities.AddError(issueContext, string(err.IssueCode()), err.Issue())
					exitOnErrors++
				},
				func(warning lp.Issue) {
					dropColl.Activities.AddWarning(issueContext, string(warning.IssueCode()), warning.Issue())
					if warning.IssueCode() == lp.MatchesIgnorePolicy {
						exitOnWarnings++
					}
				})
			if exitOnErrors > 0 || exitOnWarnings > 0 {
				ch <- index
				return
			}
		}

		finalURL, issue := link.FinalURL()
		if issue != nil {
			if issue.IsError() {
				dropColl.Activities.AddError(issueContext, string(issue.IssueCode()), issue.Issue())
			} else {
				dropColl.Activities.AddWarning(issueContext, string(issue.IssueCode()), issue.Issue())
			}
			ch <- index
			return
		}

		// this shouldnt occur because it should be caught by "issues" block above but, just in case...
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
