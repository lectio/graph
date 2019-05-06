package pipeline

import (
	"bytes"
	"fmt"
	"github.com/Machiel/slugify"
	"github.com/lectio/graph/model"
	"github.com/lectio/graph/source"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
)

// BookmarksToMarkdown converts a Bookmarks source to Hugo content
type BookmarksToMarkdown struct {
	config             *model.Configuration
	pipelineURL        *url.URL
	input              *model.BookmarksToMarkdownPipelineInput
	exec               *model.BookmarksToMarkdownPipelineExecution
	repoMan            model.RepositoryManager
	fileWriteMode      os.FileMode
	linksAPISource     model.APISource
	linksHandler       model.LinksAPIHandlerFunc
	linksHandlerParams model.LinksAPIHandlerParams
}

// NewBookmarksToMarkdown returns a new Pipeline for this strategy
func NewBookmarksToMarkdown(config *model.Configuration, input *model.BookmarksToMarkdownPipelineInput) (Pipeline, error) {
	result := new(BookmarksToMarkdown)
	result.config = config
	pipelineURL, err := url.Parse("lectio://BookmarksToMarkdown")
	if err != nil {
		return result, err
	}
	result.pipelineURL = pipelineURL
	result.input = input
	result.exec = new(model.BookmarksToMarkdownPipelineExecution)
	result.exec.Strategy = input.Strategy

	result.exec.Settings = config.SettingsBundle(input.SettingsBundle)
	if result.exec.Settings == nil {
		return result, fmt.Errorf("Unable to find settings bundle %q", input.SettingsBundle)
	}

	repoMan, err := result.exec.Settings.Repositories.OpenRepositoryName(input.Repository)
	if err != nil {
		return result, fmt.Errorf("Error opening repository %q in settings bundle %q: %v", input.Repository, input.SettingsBundle, err.Error())
	}
	result.repoMan = repoMan
	result.fileWriteMode = os.ModePerm

	result.linksAPISource, result.linksHandler, err = source.DetectAPIFromURLText(input.BookmarksURL)
	if err != nil {
		return result, err
	}
	result.linksHandlerParams, err = model.NewLinksAPIHandlerParams(config, result.linksAPISource, result.exec.Settings)
	if err != nil {
		return result, err
	}

	return result, nil
}

// IsPipelineExecution satifies model.PipelineExecution interface
func (p BookmarksToMarkdown) IsPipelineExecution() {
}

// URL is how we uniquely identify this pipeline
func (p BookmarksToMarkdown) URL() *url.URL {
	return p.pipelineURL
}

// Execute either asynchronously or synchronously runs the pipeline and returns the result
func (p *BookmarksToMarkdown) Execute() (model.PipelineExecution, error) {
	p.exec.Pipeline = model.PipelineURL(p.pipelineURL.String())
	p.exec.ExecutionID = GenerateExecutionID()
	switch p.input.Strategy {
	case model.PipelineExecutionStrategyAsynchronous:
		//TODO: start as go-routine after testing ... go p.execute()
		p.execute()
	case model.PipelineExecutionStrategySynchronous:
		p.execute()
	default:
		return p.exec, fmt.Errorf("The execution strategy should be either async or sync, not %q", p.input.Strategy.String())
	}
	return p.exec, nil
}

func (p *BookmarksToMarkdown) execute() {
	bookmarks, err := p.linksHandler(p.linksHandlerParams)
	if err != nil {
		p.exec.Activities.AddError(p.pipelineURL.String(), "BM2MDERR_LINKSHANDLER", fmt.Sprintf("Unable to retrieve bookmarks: %v", err.Error()))
		return
	}
	if bookmarks == nil {
		p.exec.Activities.AddError(p.pipelineURL.String(), "BM2MDERR_LINKSHANDLER", "Links handler did not return an error, but bookmarks is nil")
		return
	}
	p.exec.Bookmarks = bookmarks

	fs := p.repoMan.FileSystem()
	p.exec.Activities.AddHistory(&model.ActivityLog{Message: model.ActivityHumanMessage(fmt.Sprintf("Created FileSystem() +%v", fs))})

	for index, bookmark := range bookmarks.Content {
		context := fmt.Sprintf("[%q] bookmark %d", p.pipelineURL.String(), index)

		if len(p.exec.Activities.Errors) > p.input.CancelOnWriteErrors {
			p.exec.Activities.AddError(context, "BM2MDERR_WRITE_ERRORS_LIMIT_REACHED", fmt.Sprintf("Write errors limit exceeded: %d", p.input.CancelOnWriteErrors))
			return
		}

		slug := slugify.Slugify(string(bookmark.Title))

		frontmatter := make(map[string]interface{})
		frontmatter["slug"] = slug
		frontmatter["title"] = bookmark.Title
		frontmatter["description"] = bookmark.Summary

		bookmark.Properties.ForEach(func(key model.PropertyName, value interface{}) {
			_, found := frontmatter[string(key)]
			if !found {
				frontmatter[string(key)] = value
			} else {
				p.exec.Activities.AddWarning(context, "BM2MDERR_FMKEY_MERGE_DUPLICATE", fmt.Sprintf("Property name %q is duplicated, retaining earliest value", key))
			}
		})

		fmBytes, fmErr := yaml.Marshal(frontmatter)
		if fmErr != nil {
			p.exec.Activities.AddError(context, "BM2MDERR_MARSHAL_FM", fmt.Sprintf("Unable to marshal front matter: %v", fmErr.Error()))
			continue
		}
		markdown := bytes.NewBufferString("---\n")
		_, writeErr := markdown.Write(fmBytes)
		if writeErr != nil {
			p.exec.Activities.AddError(context, "BM2MDERR_WRITE_FM", fmt.Sprintf("Unable to write front matter: %v", writeErr.Error()))
			continue
		}
		_, writeErr = markdown.WriteString("---\n")
		_, writeErr = markdown.WriteString(string(bookmark.Body))
		if writeErr != nil {
			p.exec.Activities.AddError(context, "BM2MDERR_WRITE_BODY", fmt.Sprintf("Unable to write content body: %v", writeErr.Error()))
			continue
		}

		fileName := fmt.Sprintf("%s.md", slug)
		writeErr = afero.WriteFile(fs, fileName, markdown.Bytes(), p.fileWriteMode)
		if writeErr != nil {
			p.exec.Activities.AddError(context, "BM2MDERR_WRITE_MD", fmt.Sprintf("Unable to write markdown content: %v", writeErr.Error()))
			continue
		}
	}
}
