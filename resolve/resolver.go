package resolve

import (
	"context"
	"fmt"
	"github.com/lectio/graph/pipeline"

	"github.com/lectio/graph/source"

	"github.com/lectio/graph/model"
)

// Resolver is the primary Lectio Graph resolver
type Resolver struct {
	config *model.Configuration
}

// MakeResolver creates the default resolver
func MakeResolver() *Resolver {
	result := new(Resolver)
	result.config, _ = model.MakeConfiguration()
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

func (r *queryResolver) AllSettings(ctx context.Context) ([]model.PersistentSettings, error) {
	return r.config.AllSettings()
}

func (r *queryResolver) Settings(ctx context.Context, settings model.SettingsPath) ([]model.PersistentSettings, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

func (r *queryResolver) Source(ctx context.Context, urlText model.URLText) (model.ContentSource, error) {
	source, _, srcErr := source.DetectFromURLText(urlText)
	return source, srcErr
}

func (r *queryResolver) Bookmarks(ctx context.Context, sourceURL model.URLText, settings model.SettingsPath) (*model.Bookmarks, error) {
	apiSource, handler, srcErr := source.DetectAPIFromURLText(sourceURL)
	if srcErr != nil {
		return nil, srcErr
	}

	params, paramsErr := source.NewLinksAPIHandlerParams(r.config, apiSource, settings)
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
