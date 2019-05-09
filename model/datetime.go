package model

import (
	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/araddon/dateparse"
	"io"
	"time"
)

const dateFormat = "Mon Jan 2 15:04:05 MST 2006"

type DateTime time.Time

func (t DateTime) MarshalGQL(w io.Writer) {
	result := time.Time(t).Format(dateFormat)
	graphql.MarshalString(result).MarshalGQL(w)
}

func (t *DateTime) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}
	parsed, err := dateparse.ParseAny(str)
	if err != nil {
		return err
	}
	*t = DateTime(parsed)
	return nil
}
