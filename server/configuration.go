package main

import (
	"reflect"

	"github.com/pkg/errors"
)

type configuration struct {
	DisableDMOnJoin          bool     `json:"disable_dm_on_join"`
	DisableGMOnJoin          bool     `json:"disable_gm_on_join"`
	DisableDMForExistingUser bool     `json:"disable_dm_for_existing_user"`
	DisableGMForExistingUser bool     `json:"disable_gm_for_existing_user"`
	ExcludedUsers            []string `json:"excluded_users"`
	AllowedRoles             []string `json:"allowed_roles"`
}

func (c *configuration) Clone() *configuration {
	var clone = *c
	clone.ExcludedUsers = make([]string, len(c.ExcludedUsers))
	copy(clone.ExcludedUsers, c.ExcludedUsers)
	clone.AllowedRoles = make([]string, len(c.AllowedRoles))
	copy(clone.AllowedRoles, c.AllowedRoles)
	return &clone
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration.Clone()
}

func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	return nil
}
