package model

import (
	"errors"
	io "io"
	"regexp"

	graphql "github.com/99designs/gqlgen/graphql"
	"gopkg.in/jdkato/prose.v2"
)

var sourceNameAfterPipeRegEx = regexp.MustCompile(` \| .*$`)   // Matches " | Healthcare IT News" from a title like "xyz title | Healthcare IT News"
var sourceNameAfterHyphenRegEx = regexp.MustCompile(` \- .*$`) // Matches " - Healthcare IT News" from a title like "xyz title - Healthcare IT News"
var firstSentenceRegExp = regexp.MustCompile(`^(.*?)[.?!]`)

type ContentTitleText string
type ContentBodyText string
type ContentSummaryText string

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

func (t ContentTitleText) Edit(editPipedSuffix bool) (ContentTitleText, error) {
	if editPipedSuffix {
		return ContentTitleText(sourceNameAfterPipeRegEx.ReplaceAllString(string(t), "")), nil
	}
	return t, nil
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

func (t ContentSummaryText) Edit(obj *CuratedLink, firstSentenceOfBodyIfEmpty bool) (ContentSummaryText, error) {
	if len(t) == 0 && firstSentenceOfBodyIfEmpty {
		firstSentence, fsErr := obj.Body.FirstSentence()
		return ContentSummaryText(firstSentence), fsErr
	}
	return t, nil
}
