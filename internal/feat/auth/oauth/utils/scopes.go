package oauth

import (
	"slices"

	"golang.org/x/oauth2"
)

type Scopes struct {
	scopesSlice []string
}

// Array returns a copy of the underlying scopes slice.
func (s Scopes) Array() []string {
	return slices.Clone(s.scopesSlice)
}

// Scopes count
func (s Scopes) Len() int {
	return len(s.scopesSlice)
}

// Equal reports whether two Scopes are equal.
func (s Scopes) Equal(o Scopes) bool {
	return slices.Equal(s.scopesSlice, o.scopesSlice)
}

// EqualSlice reports whether the Scopes are equal to the given slice.
func (s Scopes) EqualSlice(o []string) bool {
	return s.Equal(*NewScopes(o))
}

// NewScopes constructs a Scopes instance from a slice of strings.
// This mirrors the database trigger (oauth_connection_sort_and_dedupe_scopes_array_fn),
// which enforces the same ordering and uniqueness.
// Always construct Scopes through NewScopes to guarantee equality checks
// match the database representation.
func NewScopes(s []string) *Scopes {
	scopesSlice := slices.Clone(s)
	slices.Sort(scopesSlice) // DO NOT REMOVE THIS LINE!
	return &Scopes{
		scopesSlice: scopesSlice,
	}
}

func NewScopesFromOauthToken(oauthToken *oauth2.Token) *Scopes {
	if scopes, ok := oauthToken.Extra("scope").([]string); ok {
		return NewScopes(scopes)
	}
	return NewScopes([]string{})
}
