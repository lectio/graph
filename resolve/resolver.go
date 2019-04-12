package resolve

import (
	"context"
	"time"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct{}

// HarvestedLink has a special resolver since some fields have arguments
func (r *Resolver) HarvestedLink() HarvestedLinkResolver {
	return &harvestedLinkResolver{r}
}

// Query is the the central location for all query resolvers
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type harvestedLinkResolver struct{ *Resolver }

func (r *harvestedLinkResolver) Title(ctx context.Context, obj *model.HarvestedLink, options []model.ContentTitleOption) (model.ContentTitleText, error) {
	return obj.Title.Edit(obj, options)
}
func (r *harvestedLinkResolver) Summary(ctx context.Context, obj *model.HarvestedLink, options []model.ContentSummaryOption) (model.ContentSummaryText, error) {
	return obj.Summary.Edit(obj, options)
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) DefaultSettingsBundle(ctx context.Context) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.DefaultBundle(), nil
}

func (r *queryResolver) SettingsBundle(ctx context.Context, name model.SettingsBundleName) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.SettingsBundle(name), nil
}

func (r *queryResolver) HarvestedLinks(ctx context.Context, feedURL model.URLText, settingsBundle model.SettingsBundleName) (*model.HarvestedLinks, error) {
	settings, sbErr := r.SettingsBundle(ctx, settingsBundle)
	if sbErr != nil {
		return nil, sbErr
	}

	dc, dcErr := dropmark.GetCollection(string(feedURL), nil, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	dropColl := model.HarvestedLinks{}
	for _, item := range dc.Items {
		cl := model.HarvestedLink{
			ID:      "test",
			Title:   model.ContentTitleText(item.Name),
			Summary: model.ContentSummaryText(item.Description),
			Body:    model.ContentBodyText(item.Content)}
		dropColl.Content = append(dropColl.Content, cl)
	}
	return &dropColl, nil
}
