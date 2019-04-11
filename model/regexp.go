package model

import (
	"fmt"
	io "io"
	"regexp"

	graphql "github.com/99designs/gqlgen/graphql"
)

// RegularExpression is a golang regexp.Regexp wrapped for use by GraphQL
type RegularExpression regexp.Regexp

// MakeRegularExpression creates an instance of a regular expression from a string
func MakeRegularExpression(expression string) (*RegularExpression, error) {
	result := RegularExpression{}
	reErr := result.UnmarshalGQL(expression)
	return &result, reErr
}

// MarshalGQL will marshall a regexp.Regexp to a GraphQL output stream
func (t RegularExpression) MarshalGQL(w io.Writer) {
	re := regexp.Regexp(t)
	graphql.MarshalString(re.String()).MarshalGQL(w)
}

// UnmarshalGQL will take a GraphQL string and convert it to a regular expression
func (t *RegularExpression) UnmarshalGQL(v interface{}) error {
	expression, strErr := graphql.UnmarshalString(v)
	if strErr != nil {
		return strErr
	}
	if expression != "" {
		re, reErr := regexp.Compile(expression)
		if reErr != nil {
			return fmt.Errorf(`regexp '%s' is invalid: %s`, expression, reErr)
		}
		*t = RegularExpression(*re)
	}
	return nil
}

// MatchString returns true if text matches the regular expression
func (t RegularExpression) MatchString(text string) bool {
	re := regexp.Regexp(t)
	return re.MatchString(text)
}

// String returns the text version of the regular expression
func (t RegularExpression) String() string {
	re := regexp.Regexp(t)
	return re.String()
}
