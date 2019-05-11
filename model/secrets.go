package model

import (
	graphql "github.com/99designs/gqlgen/graphql"
	"github.com/lectio/score"
	"github.com/lectio/secret"
	"io"
)

// SecretsVault manages symmetric encryption
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

// SharedCountAPIKey satisfies score.SharedCountCredentials interface for getting the SharedCount.com service credentials
func (v SecretsVault) SharedCountAPIKey() (string, bool, score.Issue) {
	apiKey, err := v.vault.DecryptText("0d4af7674abbfa18d01510fc107318ace74175c5cae32b1e3dfb1ec37ee5ceb1c8253d880ba027ed3c8280883cef0152d447f068a21f0a793f83c552fd89703aeecd53d5")
	if err != nil {
		return "", false, score.NewIssue("SharedCount.com", score.SecretManagementError, err.Error(), true)
	}
	return apiKey, true, nil
}

// EncryptText encrypts text
func (s *SecretText) EncryptText(text string) (string, error) {
	var err error
	s.EncryptedText, err = s.Vault.vault.EncryptText(text)
	return s.EncryptedText, err
}

// DecryptText decrypts
func (s SecretText) DecryptText() (string, error) {
	return s.Vault.vault.DecryptText(s.EncryptedText)
}
