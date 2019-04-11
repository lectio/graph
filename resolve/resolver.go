package resolve

import (
	"context"
	"fmt"
	"time"

	"github.com/lectio/dropmark"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct{}

// CuratedLink has a special resolver since some fields have arguments
func (r *Resolver) CuratedLink() CuratedLinkResolver {
	return &curatedLinkResolver{r}
}

// Query is the the central location for all query resolvers
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type curatedLinkResolver struct{ *Resolver }

func (r *curatedLinkResolver) Title(ctx context.Context, obj *model.CuratedLink, editPipedSuffix bool) (model.ContentTitleText, error) {
	return obj.Title.Edit(editPipedSuffix)
}

func (r *curatedLinkResolver) Summary(ctx context.Context, obj *model.CuratedLink, firstSentenceOfBodyIfEmpty bool) (model.ContentSummaryText, error) {
	return obj.Summary.Edit(obj, firstSentenceOfBodyIfEmpty)
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) DefaultSettingsBundle(ctx context.Context) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.DefaultBundle(), nil
}

func (r *queryResolver) SettingsBundle(ctx context.Context, name model.SettingsBundleName) (*model.SettingsBundle, error) {
	return model.GlobalConfiguration.SettingsBundle(name), nil
}

func (r *queryResolver) DropmarkCollection(ctx context.Context, feedID string, settingsBundle model.SettingsBundleName) (*model.DropmarkCollection, error) {
	settings, sbErr := r.SettingsBundle(ctx, settingsBundle)
	if sbErr != nil {
		return nil, sbErr
	}

	dropmarkURL := fmt.Sprintf("https://shah.dropmark.com/%s.json", feedID)
	dc, dcErr := dropmark.GetCollection(dropmarkURL, nil, settings.HTTPClient.UserAgent, time.Duration(settings.HTTPClient.Timeout))
	if dcErr != nil {
		return nil, dcErr
	}

	dropColl := model.DropmarkCollection{}
	for _, item := range dc.Items {
		cl := model.CuratedLink{
			ID:      "test",
			Title:   model.ContentTitleText(item.Name),
			Summary: model.ContentSummaryText(item.Description),
			Body:    model.ContentBodyText(item.Content)}
		dropColl.Content = append(dropColl.Content, cl)
	}
	return &dropColl, nil
}
