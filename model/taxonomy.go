package model

type TaxonomyName string
type TaxonName string

func (t *FlatTaxonomy) Add(name TaxonName) {
	t.Taxa = append(t.Taxa, name)
}
