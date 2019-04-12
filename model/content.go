package model

import (
	"errors"
	"fmt"
	io "io"
	"regexp"

	"github.com/lectio/frontmatter"

	graphql "github.com/99designs/gqlgen/graphql"
	"gopkg.in/jdkato/prose.v2"
)

var sourceNameAfterPipeRegEx = regexp.MustCompile(` \| .*$`)   // Matches " | Healthcare IT News" from a title like "xyz title | Healthcare IT News"
var sourceNameAfterHyphenRegEx = regexp.MustCompile(` \- .*$`) // Matches " - Healthcare IT News" from a title like "xyz title - Healthcare IT News"
var firstSentenceRegExp = regexp.MustCompile(`^(.*?)[.?!]`)

type ContentTitleText string
type ContentSummaryText string
type ContentBodyText string

func (t ContentTitleText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ContentTitleText) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ContentTitleText(str)
	}
	return err
}

func (t *ContentTitleText) Edit(obj *HarvestedLink, settings *ContentTitleSettings) error {
	if settings.RemovePipedSuffix {
		*t = ContentTitleText(sourceNameAfterPipeRegEx.ReplaceAllString(string(*t), ""))
	}
	return nil
}

func (t ContentBodyText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ContentBodyText) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ContentBodyText(str)
	}
	return err
}

func (t *ContentBodyText) Edit(obj *HarvestedLink, settings *ContentBodySettings) error {
	if settings.AllowFrontmatter {
		frontMatter := make(map[string]interface{})
		body, haveFrontMatter, fmErr := frontmatter.ParseYAMLFrontMatter([]byte(*t), frontMatter)
		if fmErr != nil {
			return fmErr
		}
		if haveFrontMatter {
			for k, v := range frontMatter {
				key := settings.FrontMatterPropertyPrefix + k
				switch value := v.(type) {
				case string:
					obj.Properties = append(obj.Properties, TextProperty{Name: key, Value: value})
				case int:
					obj.Properties = append(obj.Properties, NumericProperty{Name: key, Value: value})
				default:
					obj.Properties = append(obj.Properties, MessageProperty{Name: key, Message: fmt.Sprintf("Unknown property type for %q in frontmatter: %+v", key, v)})
				}
			}
			*t = ContentBodyText(fmt.Sprintf("%s", body))
			obj.Properties = append(obj.Properties, FlagProperty{Name: "haveFrontMatter", Value: true})
		}
	}
	return nil
}

func (t ContentBodyText) FirstSentence() (string, error) {
	return firstSentenceRegExp.FindString(string(t)), nil
}

func (t ContentBodyText) FirstSentenceNLP() (string, error) {
	content, proseErr := prose.NewDocument(string(t))
	if proseErr != nil {
		return "", proseErr
	}

	sentences := content.Sentences()
	if len(sentences) > 0 {
		return sentences[0].Text, nil
	}
	return "", errors.New("Unable to find any sentences in the body")
}

func (t ContentSummaryText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ContentSummaryText) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ContentSummaryText(str)
	}
	return err
}

func (t *ContentSummaryText) Edit(obj *HarvestedLink, settings *ContentSummarySettings) error {
	if settings.UseFirstSentenceOfBody {
		fs, _ := obj.Body.FirstSentence()
		*t = ContentSummaryText(fs)
	} else if len(*t) == 0 && settings.UseFirstSentenceOfBodyIfEmpty {
		fs, _ := obj.Body.FirstSentence()
		*t = ContentSummaryText(fs)
	}
	return nil
}
