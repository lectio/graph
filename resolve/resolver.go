package resolve

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct{}

// Query is the the central location for all query resolvers
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) DefaultSettingsBundle(ctx context.Context) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.DefaultBundle(), nil
}

func (r *queryResolver) SettingsBundle(ctx context.Context, name model.SettingsBundleName) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.SettingsBundle(name), nil
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

func (r *queryResolver) HarvestedLinks(ctx context.Context, sourceURL model.URLText, settingsBundle model.SettingsBundleName) (*model.HarvestedLinks, error) {
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

	dc, dcErr := dropmark.GetCollection(string(sourceURL), nil, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	// TODO: use graphql.CollectFieldsCtx(ctx, []string{"Content"}) or something similar to not collect fields that aren't in the selection set

	dropColl := model.HarvestedLinks{}
	for _, item := range dc.Items {
		hl := model.HarvestedLink{
			ID:         "test",
			Title:      model.ContentTitleText(item.Name),
			Summary:    model.ContentSummaryText(item.Description),
			Body:       model.ContentBodyText(item.Content),
			Properties: model.MakeProperties()}

		hl.Title.Edit(&hl, &settings.Content.Title)
		hl.Summary.Edit(&hl, &settings.Content.Summary)
		hl.Body.Edit(&hl, &settings.Content.Body)

		dropColl.Content = append(dropColl.Content, hl)
	}
	return &dropColl, nil
}
