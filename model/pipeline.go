package model

import (
	io "io"
	"strconv"

	graphql "github.com/99designs/gqlgen/graphql"
)

type PipelineURL string
type PipelineExecutionID uint64

func (t PipelineURL) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *PipelineURL) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = PipelineURL(str)
	}
	return err
}

func (t PipelineExecutionID) MarshalGQL(w io.Writer) {
	graphql.MarshalInt64(int64(t)).MarshalGQL(w)
}

func (t *PipelineExecutionID) UnmarshalGQL(v interface{}) error {
	u, err := strconv.ParseUint(v.(string), 10, 64)
	if err == nil {
		*t = PipelineExecutionID(u)
	}
	return err
}
