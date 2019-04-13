package resolve

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/lectio/link"

	lc "github.com/lectio/link/cache"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct {
	config   *model.Configuration
	linkKeys link.Keys
	pr       ProgressReporter
}

// MakeResolver creates the default resolver
func MakeResolver() *Resolver {
	result := new(Resolver)
	result.config, _ = model.MakeConfiguration()
	result.linkKeys = link.MakeDefaultKeys()
	result.pr = makeProgressReporter(true)
	return result
}

// Query is the the central location for all query resolvers
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) DefaultSettingsBundle(ctx context.Context) (*model.SettingsBundle, error) {
	return r.config.DefaultBundle(), nil
}

func (r *queryResolver) SettingsBundle(ctx context.Context, name model.SettingsBundleName) (*model.SettingsBundle, error) {
	return r.config.SettingsBundle(name), nil
}

func (r *queryResolver) Source(ctx context.Context, source model.URLText) (model.ContentSource, error) {
	result := model.APISource{}
	dropmark := regexp.MustCompile(`^https\://(.*).dropmark.com/([0-9]+).json$`)

	switch {
	case dropmark.MatchString(string(source)):
		result.Name = "Dropmark"
		result.APIEndpoint = source
	default:
		result.Name = "Unidentified"
		result.APIEndpoint = source
	}
	return result, nil
}

func (r *queryResolver) HarvestedLinks(ctx context.Context, sourceURL model.URLText, invalidLinks model.InvalidLinkPolicy, duplicateLinks model.DuplicatesRetentionPolicy, settingsBundle model.SettingsBundleName) (*model.HarvestedLinks, error) {
	source, srcErr := r.Source(ctx, sourceURL)
	if srcErr != nil {
		return nil, srcErr
	}
	switch v := source.(type) {
	case model.APISource:
		if v.Name != "Dropmark" {
			return nil, fmt.Errorf("Unkown source at %q", sourceURL)
		}
	default:
		return nil, fmt.Errorf("Unkown source at %q", sourceURL)
	}

	settings, sbErr := r.SettingsBundle(ctx, settingsBundle)
	if sbErr != nil {
		return nil, sbErr
	}
	cache, cacheErr := lc.MakeFileCache("link-cache", true, r.linkKeys, settings.Harvester, settings.Harvester, settings.Harvester)
	if cacheErr != nil {
		return nil, cacheErr
	}
	defer cache.Close()
	fmt.Println("Created cache")

	dc, dcErr := dropmark.GetCollection(string(sourceURL), r.pr, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	dropColl := model.HarvestedLinks{}
	dropColl.Source = source
	// TODO: use graphql.CollectFieldsCtx(ctx, []string{"Content"}) or something similar to not collect fields that aren't in the selection set

	work := func(ch chan<- int, index int, item *dropmark.Item) {
		hl := model.HarvestedLink{
			ID:         "test",
			URLText:    model.URLText(item.Link),
			Title:      model.ContentTitleText(item.Name),
			Summary:    model.ContentSummaryText(item.Description),
			Body:       model.ContentBodyText(item.Content),
			Properties: model.MakeProperties()}

		hl.Title.Edit(&hl, &settings.Content.Title)
		hl.Summary.Edit(&hl, &settings.Content.Summary)
		hl.Body.Edit(&hl, &settings.Content.Body)

		link, harvestErr := hl.URLText.Link(cache)
		if harvestErr == nil && link != nil {
			hl.IsValid = link.IsURLValid && link.IsDestValid
			hl.FinalizedURL = model.MakeURL(link.FinalizedURL)
			hl.IsIgnored = link.IsURLIgnored
		} else {
			hl.IsValid = false
		}
		dropColl.Content = append(dropColl.Content, hl)
		ch <- index
	}

	if r.pr != nil && r.pr.IsProgressReportingRequested() {
		r.pr.StartReportableActivity(len(dc.Items))
	}
	ch := make(chan int)
	for index, item := range dc.Items {
		go work(ch, index, item)
	}
	for range dc.Items {
		_ = <-ch
		if r.pr != nil && r.pr.IsProgressReportingRequested() {
			r.pr.IncrementReportableActivityProgress()
		}
	}
	if r.pr != nil && r.pr.IsProgressReportingRequested() {
		apiSource := dropColl.Source.(model.APISource)
		r.pr.CompleteReportableActivityProgress(fmt.Sprintf("Completed creating %d %s Links from %q", len(dropColl.Content), apiSource.Name, apiSource.APIEndpoint))
	}
	return &dropColl, nil
}
