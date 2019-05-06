package model

import (
	io "io"

	graphql "github.com/99designs/gqlgen/graphql"
)

type ActivityContext string
type ActivityCode string
type ActivityHumanMessage string
type ActivityMachineMessage string

func (a *Activities) AddError(context, code, message string) {
	a.Errors = append(a.Errors, ActivityError{
		ID:      "TODO_not_assigned_yet",
		Context: ActivityContext(context),
		Code:    ActivityCode(code),
		Message: ActivityHumanMessage(message)})
}

func (a *Activities) AddWarning(context, code, message string) {
	a.Warnings = append(a.Warnings, ActivityWarning{
		ID:      "TODO_not_assigned_yet",
		Context: ActivityContext(context),
		Code:    ActivityCode(code),
		Message: ActivityHumanMessage(message)})
}

func (a *Activities) AddHistory(activity Activity) {
	a.History = append(a.History, activity)
}

func (t ActivityContext) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ActivityContext) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ActivityContext(str)
	}
	return err
}

func (t ActivityCode) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *ActivityCode) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = ActivityCode(str)
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
