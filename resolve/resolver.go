package resolve

import (
	"context"
	"fmt"
	"github.com/lectio/graph/pipeline"

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

// Mutation is the the central location for all mutation resolvers
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
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

func (r *queryResolver) Bookmarks(ctx context.Context, sourceURL model.URLText, settingsBundleName model.SettingsBundleName) (*model.Bookmarks, error) {
	apiSource, handler, srcErr := source.DetectAPIFromURLText(sourceURL)
	if srcErr != nil {
		return nil, srcErr
	}

	settings := r.config.SettingsBundle(settingsBundleName)
	if settings == nil {
		return nil, fmt.Errorf("settings bundle %q not found in resolve.Bookmarks", settingsBundleName)
	}

	params, paramsErr := model.NewLinksAPIHandlerParams(r.config, apiSource, settings)
	if paramsErr != nil {
		return nil, paramsErr
	}

	return handler(params)
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ExecutePipeline(ctx context.Context, input model.ExecutePipelineInput) (model.PipelineExecution, error) {
	panic("not implemented")
}

func (r *mutationResolver) ExecuteBookmarksToMarkdownPipeline(ctx context.Context, input model.BookmarksToMarkdownPipelineInput) (*model.BookmarksToMarkdownPipelineExecution, error) {
	p, perr := pipeline.NewBookmarksToMarkdown(r.config, &input)
	if perr != nil {
		return nil, perr
	}
	result, err := p.Execute()
	if err != nil {
		return nil, err
	}
	return result.(*model.BookmarksToMarkdownPipelineExecution), nil
}
