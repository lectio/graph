package model

import (
	io "io"

	graphql "github.com/99designs/gqlgen/graphql"
)

type ActivityLogEntryCode string
type ActivityHumanMessage string
type ActivityMachineMessage string

func MakeActivities() *Activities {
	result := new(Activities)
	return result
}

func (t ActivityLogEntryCode) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ActivityLogEntryCode) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ActivityLogEntryCode(str)
	}
	return err
}

func (t ActivityHumanMessage) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ActivityHumanMessage) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ActivityHumanMessage(str)
	}
	return err
}

func (t ActivityMachineMessage) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ActivityMachineMessage) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ActivityMachineMessage(str)
	}
	return err
}
