package service

import (
	"encoding/json"

	"github.com/fleetdm/fleet/v4/server/fleet"
)

// ListTeams retrieves the list of teams.
func (c *Client) ListTeams(query string) ([]fleet.Team, error) {
	verb, path := "GET", "/api/latest/fleet/teams"
	var responseBody listTeamsResponse
	err := c.authenticatedRequestWithQuery(nil, verb, path, &responseBody, query)
	if err != nil {
		return nil, err
	}
	return responseBody.Teams, nil
}

// ApplyTeams sends the list of Teams to be applied to the
// Fleet instance.
func (c *Client) ApplyTeams(specs []json.RawMessage, opts fleet.ApplySpecOptions) error {
	verb, path := "POST", "/api/latest/fleet/spec/teams"
	var responseBody applyTeamSpecsResponse
	return c.authenticatedRequestWithQuery(map[string]interface{}{"specs": specs}, verb, path, &responseBody, opts.RawQuery())
}

// ApplyPolicies sends the list of Policies to be applied to the
// Fleet instance.
func (c *Client) ApplyPolicies(specs []*fleet.PolicySpec) error {
	req := applyPolicySpecsRequest{Specs: specs}
	verb, path := "POST", "/api/latest/fleet/spec/policies"
	var responseBody applyPolicySpecsResponse
	return c.authenticatedRequest(req, verb, path, &responseBody)
}
