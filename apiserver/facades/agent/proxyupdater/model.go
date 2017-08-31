// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package proxyupdater

import (
	"github.com/juju/juju/apiserver/facade"
	"github.com/juju/juju/state"
)

// NewAPI creates a new API server-side facade with a state.State backing.
func NewAPI(st *state.State, res facade.Resources, auth facade.Authorizer) (*ProxyUpdaterAPI, error) {
	m, err := st.Model()
	if err != nil {
		return nil, err
	}
	return NewAPIWithBacking(&stateShim{st: st, m: m}, res, auth)
}
