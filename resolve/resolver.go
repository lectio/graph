package resolve

import (
	"context"
	"fmt"

	"github.com/lectio/graph/source"

	"github.com/lectio/link"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct {
	config   *model.Configuration
	linkKeys link.Keys
}

// MakeResolver creates the default resolver
func MakeResolver() *Resolver {
	result := new(Resolver)
	result.config, _ = model.MakeConfiguration()
	result.linkKeys = link.MakeDefaultKeys()
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

func (r *queryResolver) Source(ctx context.Context, urlText model.URLText) (model.ContentSource, error) {
	source, _, srcErr := source.DetectFromURLText(urlText)
	return source, srcErr
}

func (r *queryResolver) Links(ctx context.Context, sourceURL model.URLText, settingsBundleName model.SettingsBundleName) (*model.HarvestedLinks, error) {
	apiSource, handler, srcErr := source.DetectAPIFromURLText(sourceURL)
	if srcErr != nil {
		return nil, srcErr
	}

	params, paramsErr := model.NewLinksAPIHandlerParams(r.config, apiSource, settingsBundleName)
	if paramsErr != nil {
		return nil, paramsErr
	}

	fmt.Printf("%+v\n", params)

	return handler(params)
}
