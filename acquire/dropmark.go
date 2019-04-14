package acquire

import (
	"fmt"
	"time"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// DropmarkLinks returns a collection of harvested links from a Dropmark API
func DropmarkLinks(params model.LinksAPIHandlerParams) (*model.HarvestedLinks, error) {
	pr := params.ProgressReporter()
	settings := params.Settings()

	dc, dcErr := dropmark.GetCollection(string(params.Source().APIEndpoint), pr, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	dropColl := model.HarvestedLinks{}
	dropColl.Source = params.Source()
	// TODO: use graphql.CollectFieldsCtx(ctx, []string{"Content"}) or something similar to not collect fields that aren't in the selection set

	work := func(ch chan<- int, index int, item *dropmark.Item) {
		hl := model.HarvestedLink{
			ID:         settings.Harvester.PrimaryKeyForURLText(item.Link),
			URLText:    model.URLText(item.Link),
			Title:      model.ContentTitleText(item.Name),
			Summary:    model.ContentSummaryText(item.Description),
			Body:       model.ContentBodyText(item.Content),
			Properties: model.MakeProperties()}

		hl.Title.Edit(&hl, &settings.Content.Title)
		hl.Summary.Edit(&hl, &settings.Content.Summary)
		hl.Body.Edit(&hl, &settings.Content.Body)

		link, harvestErr := hl.URLText.Link(params.LinksCache())
		if harvestErr == nil && link != nil {
			hl.ID = link.PrimaryKey(settings.Harvester)
			hl.IsValid = link.IsURLValid && link.IsDestValid
			hl.FinalizedURL = model.MakeURL(link.FinalizedURL)
			hl.IsIgnored = link.IsURLIgnored
			if hl.IsIgnored && len(link.IgnoreReason) > 0 {
				hl.IgnoreReason = new(model.InterpolatedMessage)
				hl.IgnoreReason.UnmarshalGQL(link.IgnoreReason)
			}
		} else {
			hl.IsValid = false
		}
		dropColl.Content = append(dropColl.Content, hl)
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
		apiSource := dropColl.Source.(*model.APISource)
		pr.CompleteReportableActivityProgress(fmt.Sprintf("Completed creating %d %s Links from %q", len(dropColl.Content), apiSource.Name, apiSource.APIEndpoint))
	}
	return &dropColl, nil
}
