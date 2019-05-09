package model

import (
	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/lectio/secret"
	"io"
)

type SecretsVault struct {
	urlText string
	vault   secret.Vault
}

// MakeSecretsVault creates an instance of a secrets vault from a string
func MakeSecretsVault(urlText string) (*SecretsVault, error) {
	result := SecretsVault{}
	reErr := result.UnmarshalGQL(urlText)
	return &result, reErr
}

// MarshalGQL will marshall a SecretsVault to a GraphQL output stream
func (v SecretsVault) MarshalGQL(w io.Writer) {
	graphql.MarshalString(v.urlText).MarshalGQL(w)
}

// UnmarshalGQL will take a GraphQL string and convert it to a SecretsVault
func (v *SecretsVault) UnmarshalGQL(gql interface{}) error {
	urlText, err := graphql.UnmarshalString(gql)
	if err != nil {
		return err
	}
	v.urlText = urlText
	v.vault, err = secret.Parse(urlText)
	return err
}

func (s *SecretText) EncryptText(text string) (string, error) {
	var err error
	s.EncryptedText, err = s.Vault.vault.EncryptText(text)
	return s.EncryptedText, err
}

func (s SecretText) DecryptText() (string, error) {
	return s.Vault.vault.DecryptText(s.EncryptedText)
}
