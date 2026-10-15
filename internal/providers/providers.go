// Package providers contains functions to interact with different subtitle providers.
package providers

import (
	"errors"
	"iter"
)

// NewProviderSet creates a new ProviderSet
func NewProviderSet() ProviderSet {
	return ProviderSet{
		providers: make([]Provider, 0, 5),
	}
}

// Append adds one or more providers to the ProviderSet and returns the updated set.
func (set *ProviderSet) Append(provider ...Provider) *ProviderSet {
	set.providers = append(set.providers, provider...)
	return set
}

func (set *ProviderSet) Len() int {
	return len(set.providers)
}

func (set *ProviderSet) NextProvider() (Provider, error) {
	if set.nextProviderIdx >= len(set.providers) {
		return nil, ErrNoMoreProviders
	}
	provider := set.providers[set.nextProviderIdx]
	set.nextProviderIdx++
	return provider, nil
}

func (set *ProviderSet) Reset() {
	set.nextProviderIdx = 0
}

func (set *ProviderSet) Iter() iter.Seq[Provider] {
	return func(yield func(Provider) bool) {
		for {
			provider, err := set.NextProvider()
			if err != nil && errors.Is(err, ErrNoMoreProviders) {
				set.Reset()
				return
			}

			if !yield(provider) {
				return
			}
		}
	}
}
