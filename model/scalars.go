package model

import (
	io "io"
	"time"

	graphql "github.com/99designs/gqlgen/graphql"
)

type NameText string
type SmallText string
type MediumText string
type LargeText string
type ExtraLargeText string
type ErrorMessage string
type WarningMessage string
type SettingsBundleName string
type InterpolatedMessage string

type AsymmetricCryptoPublicKey string
type AsymmetricCryptoPublicKeyName string
type IdentityPrincipal string
type IdentityPassword string
type IdentityKey string

type StorageKey string

type AuthenticatedSessionID string
type AuthenticatedSessionsCount uint
type AuthenticatedSessionTimeout uint

type TimeoutDuration time.Duration

type DirectoryPath string
type FilePathAndName string
type FileNameOnly string

func (t NameText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t SmallText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t MediumText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t LargeText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *LargeText) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = LargeText(str)
	}
	return err
}

func (t ExtraLargeText) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t ErrorMessage) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t WarningMessage) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t DirectoryPath) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t SettingsBundleName) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *SettingsBundleName) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = SettingsBundleName(str)
	}
	return err
}
func (t InterpolatedMessage) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *InterpolatedMessage) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = InterpolatedMessage(str)
	}
	return err
}

func (t IdentityPrincipal) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t IdentityPassword) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t IdentityKey) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t AuthenticatedSessionID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *AuthenticatedSessionID) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = AuthenticatedSessionID(str)
	}
	return err
}

func (t AsymmetricCryptoPublicKey) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *AsymmetricCryptoPublicKey) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = AsymmetricCryptoPublicKey(str)
	}
	return err
}

func (t AsymmetricCryptoPublicKeyName) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *AsymmetricCryptoPublicKeyName) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = AsymmetricCryptoPublicKeyName(str)
	}
	return err
}

func (t AuthenticatedSessionTimeout) MarshalGQL(w io.Writer) {
	graphql.MarshalInt(int(t)).MarshalGQL(w)
}

func (t TimeoutDuration) MarshalGQL(w io.Writer) {
	graphql.MarshalString(time.Duration(t).String()).MarshalGQL(w)
}

func (t *TimeoutDuration) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		var td time.Duration
		td, err = time.ParseDuration(str)
		*t = TimeoutDuration(td)
	}
	return err
}

func (t AuthenticatedSessionsCount) MarshalGQL(w io.Writer) {
	graphql.MarshalInt(int(t)).MarshalGQL(w)
}

func (t StorageKey) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *StorageKey) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = StorageKey(str)
	}
	return err
}
