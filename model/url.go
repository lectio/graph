package model

import (
	io "io"
	"net/url"

	"github.com/lectio/link"

	graphql "github.com/99designs/gqlgen/graphql"
)

type URLText string
type URL url.URL
type Resource string

func (t URLText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *URLText) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = URLText(str)
	}
	return err
}

// SimplifiedHostname returns the URL's hostname without 'www.' prefix
func (t URLText) SimplifiedHostname() string {
	u, e := url.Parse(string(t))
	if e == nil {
		return link.GetSimplifiedHostname(u)
	}
	return string(t)
}

// SimplifiedHostnameWithoutTLD returns the URL's hostname without 'www.' prefix and removes the top level domain suffix (.com, etc.)
func (t URLText) SimplifiedHostnameWithoutTLD() string {
	u, e := url.Parse(string(t))
	if e == nil {
		return link.GetSimplifiedHostnameWithoutTLD(u)
	}
	return string(t)
}

func (t URL) MarshalGQL(w io.Writer) {
	url := url.URL(t)
	graphql.MarshalString(url.String()).MarshalGQL(w)
}

func (t *URL) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		var u *url.URL
		u, err = url.Parse(str)
		*t = URL(*u)
	}
	return err
}

// SimplifiedHostname returns the URL's hostname without 'www.' prefix
func (t URL) SimplifiedHostname() string {
	u := url.URL(t)
	return link.GetSimplifiedHostname(&u)
}

// SimplifiedHostnameWithoutTLD returns the URL's hostname without 'www.' prefix and removes the top level domain suffix (.com, etc.)
func (t URL) SimplifiedHostnameWithoutTLD() string {
	u := url.URL(t)
	return link.GetSimplifiedHostnameWithoutTLD(&u)
}

func (t Resource) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *Resource) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = Resource(str)
	}
	return err
}

func (t Resource) Harvest(cleanCurationTargetRule link.CleanResourceParamsRule, ignoreCurationTargetRule link.IgnoreResourceRule,
	followHTMLRedirect link.FollowRedirectsInCurationTargetHTMLPayload) *link.Resource {
	return link.HarvestResource(string(t), cleanCurationTargetRule, ignoreCurationTargetRule, followHTMLRedirect)
}
