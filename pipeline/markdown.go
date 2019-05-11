package pipeline

import (
	"bytes"
	"fmt"
	"github.com/Machiel/slugify"
	"github.com/lectio/graph/model"
	"github.com/lectio/graph/source"
	"github.com/lectio/image"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// BookmarksToMarkdown converts a Bookmarks source to Hugo content
type BookmarksToMarkdown struct {
	config             *model.Configuration
	settingsPath       model.SettingsPath
	pipelineURL        *url.URL
	input              *model.BookmarksToMarkdownPipelineInput
	exec               *model.BookmarksToMarkdownPipelineExecution
	repoMan            model.RepositoryManager
	fileWriteMode      os.FileMode
	linksAPISource     model.APISource
	linksHandler       model.LinksAPIHandlerFunc
	linksHandlerParams model.LinksAPIHandlerParams
	markdownSettings   *model.MarkdownGeneratorSettings
	baseFS             afero.Fs
	contentFS          afero.Fs
	imageCacheFS       afero.Fs
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

	result.settingsPath = input.Settings

	repoMan, err := config.Repositories(result.settingsPath).OpenRepositoryName(input.Repository)
	if err != nil {
		return result, fmt.Errorf("Error opening repository %q in settings path %q: %v", input.Repository, result.settingsPath, err.Error())
	}
	result.repoMan = repoMan
	result.fileWriteMode = os.ModePerm

	result.linksAPISource, result.linksHandler, err = source.DetectAPIFromURLText(input.BookmarksURL)
	if err != nil {
		return result, err
	}
	result.linksHandlerParams, err = model.NewLinksAPIHandlerParams(config, result.linksAPISource, result.settingsPath)
	if err != nil {
		return result, err
	}

	ms := config.MarkdownGeneratorSettings(result.settingsPath)
	result.baseFS = repoMan.FileSystem()
	err = result.baseFS.MkdirAll(ms.ContentPath, repoMan.DirPerm())
	if err != nil {
		return result, fmt.Errorf("Unable to create content directory %q: %v", ms.ContentPath, err.Error())
	}
	err = result.baseFS.MkdirAll(result.markdownSettings.ImagesPath, repoMan.DirPerm())
	if err != nil {
		return result, fmt.Errorf("Unable to create content directory %q: %v", ms.ImagesPath, err.Error())
	}
	result.contentFS = afero.NewBasePathFs(result.baseFS, ms.ContentPath)
	result.imageCacheFS = afero.NewBasePathFs(result.baseFS, ms.ImagesPath)

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

func (p *BookmarksToMarkdown) frontmatter(context string, bookmark *model.Bookmark) map[string]interface{} {
	apiSource := p.linksAPISource.(*model.BookmarksAPISource)
	slug := slugify.Slugify(bookmark.Link.FinalURL.BrandWithoutTLD() + "-" + string(bookmark.Title))

	frontmatter := make(map[string]interface{})
	frontmatter["archetype"] = "bookmark"
	frontmatter["source"] = apiSource
	frontmatter["date"], _ = bookmark.Properties.GetDate("dropmark.updatedAt")
	frontmatter["link"] = bookmark.Link.FinalURL.Text()
	frontmatter["linkBrand"] = bookmark.Link.FinalURL.Brand()
	frontmatter["slug"] = slug
	frontmatter["title"] = bookmark.Title
	frontmatter["description"] = bookmark.Summary

	lls := p.linksHandlerParams.LinkLifecyleSettings()
	if lls.ScoreLinks.Score {
		scores, scIssue := lls.ScoreLink(bookmark.Link.FinalURL.URL())
		if scIssue != nil {
			p.exec.Activities.AddError(scIssue.IssueContext().(string), scIssue.IssueCode(), scIssue.Issue())
		} else if scores != nil {
			frontmatter["socialScore"] = scores.SharesCount()
			if lls.ScoreLinks.Simulate {
				frontmatter["socialScoreSimulated"] = true
			}
		}
	}

	if bookmark.Taxonomies != nil && len(bookmark.Taxonomies) > 0 {
		for _, taxn := range bookmark.Taxonomies {
			switch taxonomy := taxn.(type) {
			case model.FlatTaxonomy:
				frontmatter[string(taxonomy.Name)] = taxonomy.Taxa
			case model.HiearchicalTaxonomy:
				frontmatter[string(taxonomy.Name)] = taxonomy.Taxa
			default:
				panic(fmt.Sprintf("Unknown taxonomy type %T", taxn))
			}
		}
	}

	bookmark.Properties.ForEach(func(key model.PropertyName, value interface{}) {
		_, found := frontmatter[string(key)]
		if !found {
			switch string(key) {
			case "dropmark.updatedAt":
				// skip this, the 'date' is already set
			case "dropmark.thumbnailURL":
				thumbnailURL := value.(string)
				if thumbnailURL != "" {
					fileName, _, issue := image.Cache(thumbnailURL, p, slug)
					if issue == nil {
						frontmatter["featuredImage"] = fmt.Sprintf("%s/%s", p.markdownSettings.ImagesURLRel, fileName)
					} else {
						frontmatter["featuredImageCacheErr"] = fmt.Sprintf("[%q] [%s]: %s", thumbnailURL, issue.IssueCode(), issue.Issue())
						p.exec.Activities.AddError(issue.IssueContext().(string), issue.IssueCode(), issue.Issue())
					}
				}
			default:
				frontmatter[string(key)] = value
			}
		} else {
			p.exec.Activities.AddWarning(context, "BM2MDERR_FMKEY_MERGE_DUPLICATE", fmt.Sprintf("Property name %q is duplicated, retaining earliest value", key))
		}
	})

	return frontmatter
}

func (p *BookmarksToMarkdown) write(contentFS afero.Fs, context string, bookmark *model.Bookmark, frontmatter map[string]interface{}) {
	fmBytes, fmErr := yaml.Marshal(frontmatter)
	if fmErr != nil {
		p.exec.Activities.AddError(context, "BM2MDERR_MARSHAL_FM", fmt.Sprintf("Unable to marshal front matter: %v", fmErr.Error()))
		return
	}
	markdown := bytes.NewBufferString("---\n")
	_, writeErr := markdown.Write(fmBytes)
	if writeErr != nil {
		p.exec.Activities.AddError(context, "BM2MDERR_WRITE_FM", fmt.Sprintf("Unable to write front matter: %v", writeErr.Error()))
		return
	}
	_, writeErr = markdown.WriteString("---\n")
	_, writeErr = markdown.WriteString(string(bookmark.Body))
	if writeErr != nil {
		p.exec.Activities.AddError(context, "BM2MDERR_WRITE_BODY", fmt.Sprintf("Unable to write content body: %v", writeErr.Error()))
		return
	}

	fileName := fmt.Sprintf("%s.md", frontmatter["slug"])
	writeErr = afero.WriteFile(contentFS, fileName, markdown.Bytes(), p.fileWriteMode)
	if writeErr != nil {
		p.exec.Activities.AddError(context, "BM2MDERR_WRITE_MD", fmt.Sprintf("Unable to write markdown content: %v", writeErr.Error()))
		return
	}
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

	var written uint
	pr := p.config.ProgressReporter()
	pr.StartReportableActivity(len(bookmarks.Content))
	for index, bookmark := range bookmarks.Content {
		context := fmt.Sprintf("[%q] bookmark %d", p.pipelineURL.String(), index)

		if len(p.exec.Activities.Errors) > p.markdownSettings.CancelOnWriteErrors {
			p.exec.Activities.AddError(context, "BM2MDERR_WRITE_ERRORS_LIMIT_REACHED", fmt.Sprintf("Write errors limit exceeded: %d", p.markdownSettings.CancelOnWriteErrors))
			break
		}

		frontmatter := p.frontmatter(context, &bookmark)
		p.write(p.contentFS, context, &bookmark, frontmatter)
		pr.IncrementReportableActivityProgress()
		written++
	}
	pr.CompleteReportableActivityProgress(fmt.Sprintf("Wrote %d of %d bookmarks to %+v", written, len(bookmarks.Content), p.contentFS))
}

// FileSystem satisfies image.CacheStrategy interface
func (p BookmarksToMarkdown) FileSystem() afero.Fs {
	return p.imageCacheFS
}

// HTTPUserAgent satisfies image.CacheStrategy interface
func (p BookmarksToMarkdown) HTTPUserAgent() string {
	return image.HTTPUserAgent
}

// HTTPTimeout satisfies image.CacheStrategy interface
func (p BookmarksToMarkdown) HTTPTimeout() time.Duration {
	return image.HTTPTimeout
}

// FileName satisfies image.CacheStrategy interface
func (p BookmarksToMarkdown) FileName(url *url.URL, suggested string) (string, bool) {
	extn := filepath.Ext(url.Path)
	name := fmt.Sprintf("%s%s", suggested, extn)
	found, _ := afero.Exists(p.imageCacheFS, name)
	return name, !found
}
