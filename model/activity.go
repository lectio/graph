package model

import (
	"fmt"
	io "io"
	"os"
	"path/filepath"

	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/olekukonko/tablewriter"
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

func (a *Activities) WriteMarkdown(frontmatter map[string]interface{}, activities Activities, w io.Writer) error {
	if len(frontmatter) > 0 {
		if _, err := fmt.Fprintln(w, "---"); err != nil {
			return err
		}
		for key, value := range frontmatter {
			if _, err := fmt.Fprintf(w, "%s: %s", key, value); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "---"); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w, "# Errors"); err != nil {
		return err
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Context", "Code", "Message"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, entry := range activities.Errors {
		table.Append([]string{string(entry.Context), string(entry.Code), string(entry.Message)})
	}
	table.Render()

	if _, err := fmt.Fprintln(w, "\n# Warnings"); err != nil {
		return err
	}

	table = tablewriter.NewWriter(w)
	table.SetHeader([]string{"Context", "Code", "Message"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, entry := range activities.Warnings {
		table.Append([]string{string(entry.Context), string(entry.Code), string(entry.Message)})
	}
	table.Render()

	if _, err := fmt.Fprintln(w, "\n\n# Log"); err != nil {
		return err
	}
	// for _, entry := range activities.History {
	// 	// if _, err := fmt.Fprintf(w, "\n* [%s] %s %s", entry.Context, entry.Code, entry.Message); err != nil {
	// 	// 	return err
	// 	// }
	// }

	return nil
}

func (a *Activities) WriteMarkdownFile(frontmatter map[string]interface{}, activities Activities, path string, name string) error {
	if dir, err := filepath.Abs(path); err == nil {
		if err := os.MkdirAll(dir, os.FileMode(0755)); err == nil {
			if file, err := os.Create(filepath.Join(dir, name)); err == nil {
				defer file.Close()
				return a.WriteMarkdown(frontmatter, activities, file)
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
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
