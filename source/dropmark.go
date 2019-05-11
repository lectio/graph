package source

import (
	"fmt"
	"sync"
	"time"

	"github.com/lectio/dropmark"
	ll "github.com/lectio/link"

	"github.com/lectio/graph/model"
)

// NewBookmarkFromDropmarkLink uses the dropmark.Item to create a model.Bookmark
func NewBookmarkFromDropmarkLink(item *dropmark.Item, lm *LinksManager, cs *model.ContentSettings, errorFn func(code, message string), warnFn func(code, message string)) *model.Bookmark {
	bookmark := model.Bookmark{
		ID:         lm.PrimaryKeyForURLText(item.Link),
		Link:       model.BookmarkLink{OriginalURLText: model.URLText(item.Link)},
		Title:      model.ContentTitleText(item.Name),
		Summary:    model.ContentSummaryText(item.Description),
		Body:       model.ContentBodyText(item.Content),
		Properties: model.MakeProperties()}

	bookmark.Title.Edit(&bookmark, &cs.Title)
	bookmark.Summary.Edit(&bookmark, &cs.Summary)
	bookmark.Body.Edit(&bookmark, &cs.Body)

	if warnFn != nil && bookmark.Link.OriginalURLText.IsEmpty() {
		warnFn("DLWARN-0101-LINKEMPTY", "Empty link")
		return nil
	}

	link, linkErr := bookmark.Link.OriginalURLText.Link(lm)
	if errorFn != nil && (linkErr != nil || link == nil) {
		errorFn("DLERR-0101-LINKERR", fmt.Sprintf("Unable to create link.Link: %v", linkErr))
		return nil
	}
	managedLink, isManagedLink := link.(ll.ManagedLink)

	if isManagedLink && managedLink.Issues() != nil {
		managedLink.Issues().HandleIssues(
			func(err ll.Issue) {
				if errorFn != nil {
					errorFn(string(err.IssueCode()), err.Issue())
				}
			},
			func(warning ll.Issue) {
				if warnFn != nil {
					warnFn(string(warning.IssueCode()), warning.Issue())
				}
			})
	}

	finalURL, finalURLErr := link.FinalURL()
	if errorFn != nil && finalURLErr != nil {
		errorFn("DMERR-LINK_FINALURL", finalURLErr.Error())
		return nil
	}

	// this shouldnt occur because it should be caught by "issues" block above but, just in case...
	if warnFn != nil && isManagedLink {
		ignore, ignoreReason := managedLink.Ignore()
		if ignore {
			warnFn("DLWARN-0100-IGNORE", ignoreReason)
			return nil
		}
	}

	bookmark.ID = lm.PrimaryKeyForURL(finalURL)
	bookmark.Link.IsValid = true
	bookmark.Link.FinalURL = model.MakeURL(finalURL)

	if item.Tags != nil && len(item.Tags) > 0 {
		categories := model.FlatTaxonomy{Name: "categories"}
		for _, tag := range item.Tags {
			categories.Add(model.TaxonName(tag.Name))
		}
		bookmark.Taxonomies = append(bookmark.Taxonomies, categories)
	}

	bookmark.Properties.Add("dropmark.editURL", item.DropmarkEditURL)
	bookmark.Properties.Add("dropmark.updatedAt", item.UpdatedAt)
	bookmark.Properties.Add("dropmark.thumbnailURL", item.ThumbnailURL)

	return &bookmark
}

// DropmarkLinks returns a collection of harvested links from a Dropmark API
func DropmarkLinks(params LinksAPIHandlerParams) (*model.Bookmarks, error) {
	source, ok := params.Source().(*model.BookmarksAPISource)
	if !ok {
		return nil, fmt.Errorf("Source is %+v, worker.DropmarkLinks requires a model.BookmarksAPISource", params.Source())
	}

	pr := params.ProgressReporter()
	hcs := params.HTTPClientSettings()
	lm := params.LinksManager()
	cs := params.ContentSettings()

	dropColl := model.Bookmarks{}
	dropColl.Source = *source
	dropColl.Activities = model.Activities{}

	dropCollMutex := sync.RWMutex{}
	asynch := params.Asynch()

	dc, issues := dropmark.GetCollection(string(source.APIEndpoint), pr, hcs.UserAgent, time.Duration(hcs.Timeout))
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

	createBookmark := func(index int, item *dropmark.Item) bool {
		issueContext := fmt.Sprintf("[%s] Dropmark link %d %q", source.APIEndpoint, index, item.Link)
		bookmark := NewBookmarkFromDropmarkLink(item, lm, cs,
			func(code, message string) {
				if asynch {
					dropCollMutex.Lock()
					dropColl.Activities.AddError(issueContext, code, message)
					dropCollMutex.Unlock()
				} else {
					dropColl.Activities.AddError(issueContext, code, message)
				}
			},
			func(code, message string) {
				if asynch {
					dropCollMutex.Lock()
					dropColl.Activities.AddWarning(issueContext, code, message)
					dropCollMutex.Unlock()
				} else {
					dropColl.Activities.AddWarning(issueContext, code, message)
				}
			})

		if bookmark != nil {
			if asynch {
				dropCollMutex.Lock()
				dropColl.Content = append(dropColl.Content, *bookmark)
				dropCollMutex.Unlock()
			} else {
				dropColl.Content = append(dropColl.Content, *bookmark)
			}
			return true
		}
		return false
	}

	pr.StartReportableActivity(len(dc.Items))
	if asynch {
		var wg sync.WaitGroup
		queue := make(chan int)
		for index, item := range dc.Items {
			wg.Add(1)
			go func(index int, item *dropmark.Item) {
				defer wg.Done()
				createBookmark(index, item)
				queue <- index
			}(index, item)
		}
		go func() {
			defer close(queue)
			wg.Wait()
		}()
		for range queue {
			pr.IncrementReportableActivityProgress()
		}
	} else {
		for index, item := range dc.Items {
			createBookmark(index, item)
			pr.IncrementReportableActivityProgress()
		}
	}
	pr.CompleteReportableActivityProgress(fmt.Sprintf("Imported %d of %d %s Links from %q", len(dropColl.Content), len(dc.Items), source.Name, source.APIEndpoint))
	return &dropColl, nil
}
