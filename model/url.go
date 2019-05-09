package model

import (
	io "io"
	"net/url"

	"github.com/lectio/link"

	graphql "github.com/99designs/gqlgen/graphql"
)

type URLText string
type URL url.URL

// IsEmpty returns true if the string is blank
func (t URLText) IsEmpty() bool {
	return len(string(t)) == 0
}

func (t URLText) Link(lc link.Lifecycle) (link.Link, link.Issue) {
	return lc.HarvestLink(string(t))
}

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

func MakeURL(url *url.URL) *URL {
	if url == nil {
		return nil
	}

	t := new(URL)
	*t = URL(*url)
	return t
}

func (t URL) URL() *url.URL {
	u := url.URL(t)
	return &u
}

func (t URL) Text() string {
	url := url.URL(t)
	return url.String()
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
func (t URL) Brand() string {
	u := url.URL(t)
	return link.GetSimplifiedHostname(&u)
}

// SimplifiedHostnameWithoutTLD returns the URL's hostname without 'www.' prefix and removes the top level domain suffix (.com, etc.)
func (t URL) BrandWithoutTLD() string {
	u := url.URL(t)
	return link.GetSimplifiedHostnameWithoutTLD(&u)
}
