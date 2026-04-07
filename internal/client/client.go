package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ApiResponse struct {
	Data    json.RawMessage  `json:"data"`
	Success bool             `json:"success"`
	Error   bool             `json:"error"`
	Message string           `json:"message"`
	Status  int              `json:"status"`
	Stack   *json.RawMessage `json:"stack,omitempty"`
}

type APIError struct {
	StatusCode  int
	ApiResponse ApiResponse
}

func (e *APIError) Error() string {
	// TODO add stack here
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.ApiResponse.Message)
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiResp ApiResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("API wrong StatusCode : %d\nAPI wrong Response Type : %s", resp.StatusCode, string(respBody))
		}

		return nil, &APIError{
			StatusCode:  resp.StatusCode,
			ApiResponse: apiResp,
		}
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("API wrong ResponseType : %s", string(respBody))
	}

	if !apiResp.Success || apiResp.Error {
		return nil, &APIError{
			StatusCode:  resp.StatusCode,
			ApiResponse: apiResp,
		}
	}

	return apiResp.Data, nil
}

// Role definitions
type Role struct {
	ID          int    `json:"roleId,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateRole(orgID string, role *Role) (*Role, error) {
	path := fmt.Sprintf("/org/%s/role", orgID)
	body := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
	}
	data, err := c.doRequest("PUT", path, body)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetRole(orgID string, roleID int) (*Role, error) {
	path := fmt.Sprintf("/role/%d", roleID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateRole(orgID string, roleID int, role *Role) (*Role, error) {
	path := fmt.Sprintf("/role/%d", roleID)
	body := map[string]interface{}{
		"name":        role.Name,
		"description": role.Description,
	}
	data, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	var out Role
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteRole(orgID string, roleID int) error {
	path := fmt.Sprintf("/role/%d", roleID)
	// Workaround: Pangolin requires a replacement role ID for users in the deleted role.
	// We use ID 2 (Member) which is standard in a fresh org.
	body := map[string]interface{}{
		"roleId": "2",
	}
	_, err := c.doRequest("DELETE", path, body)
	return err
}

func (c *Client) ListRoles(orgID string) ([]Role, error) {
	path := fmt.Sprintf("/org/%s/roles", orgID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Roles []Role `json:"roles"`
	}
	err = json.Unmarshal(data, &wrapper)
	return wrapper.Roles, err
}

// Site definitions
type Site struct {
	ID      int     `json:"siteId,omitempty"`
	Name    string  `json:"name"`
	NewtID  *string `json:"newtId,omitempty"`
	Secret  *string `json:"secret,omitempty"`
	Address *string `json:"address,omitempty"`
	Subnet  *string `json:"subnet,omitempty"`
	Type    *string `json:"type,omitempty"`
}

func (c *Client) ListSites(orgID string) ([]Site, error) {
	path := fmt.Sprintf("/org/%s/sites", orgID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Sites []Site `json:"sites"`
	}
	err = json.Unmarshal(data, &wrapper)
	return wrapper.Sites, err
}

func (c *Client) CreateSite(orgID string, site Site) (*Site, error) {
	path := fmt.Sprintf("/org/%s/site", orgID)
	data, err := c.doRequest("PUT", path, site)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetSite(siteID int) (*Site, error) {
	path := fmt.Sprintf("/site/%d", siteID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateSite(siteID int, site Site) (*Site, error) {
	path := fmt.Sprintf("/site/%d", siteID)
	site.Address = nil
	site.Subnet = nil
	site.NewtID = nil
	site.Secret = nil
	site.Type = nil
	data, err := c.doRequest("POST", path, site)
	if err != nil {
		return nil, err
	}
	var out Site
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteSite(siteID int) error {
	path := fmt.Sprintf("/site/%d", siteID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// SiteResource definitions
type SiteResource struct {
	ID                 int64    `json:"siteResourceId,omitempty"`
	NiceID             string   `json:"niceId,omitempty"`
	Name               string   `json:"name"`
	Mode               string   `json:"mode"`
	SiteID             int64    `json:"siteId"`
	Destination        string   `json:"destination"`
	Enabled            *bool    `json:"enabled,omitempty"`
	Alias              *string  `json:"alias,omitempty"`
	UserIDs            []string `json:"userIds"`
	RoleIDs            []int    `json:"roleIds"`
	ClientIDs          []int    `json:"clientIds"`
	TCPPortRangeString string   `json:"tcpPortRangeString,omitempty"`
	UDPPortRangeString string   `json:"udpPortRangeString,omitempty"`
	DisableIcmp        *bool    `json:"disableIcmp,omitempty"`
}

func (c *Client) CreateSiteResource(orgID string, res *SiteResource) (*SiteResource, error) {
	path := fmt.Sprintf("/org/%s/site-resource", orgID)
	data, err := c.doRequest("PUT", path, res)
	if err != nil {
		return nil, err
	}
	var out SiteResource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetSiteResource(orgID string, siteID int64, resID int64) (*SiteResource, error) {
	// TODO use /site-resource/{siteResourceId} when this issue get resolved :
	// https://github.com/fosrl/pangolin/issues/2743
	limit := 1000
	for offset, nResources := 0, limit; nResources == limit; offset += limit {
		params := url.Values{
			"limit":  []string{strconv.Itoa(limit)},
			"offset": []string{strconv.Itoa(offset)},
		}
		u := url.URL{
			Path:     fmt.Sprintf("/org/%s/site/%d/resources", orgID, siteID),
			RawQuery: params.Encode(),
		}
		data, err := c.doRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}

		var out struct {
			SiteResources []SiteResource `json:"siteResources"`
		}
		err = json.Unmarshal(data, &out)
		if err != nil {
			return nil, err
		}

		nResources = len(out.SiteResources)
		for _, siteResource := range out.SiteResources {
			if siteResource.ID == resID {
				siteResource.RoleIDs, err = c.GetSiteResourceRoles(resID)
				if err != nil {
					return nil, err
				}
				siteResource.UserIDs, err = c.GetSiteResourceUsers(resID)
				if err != nil {
					return nil, err
				}
				siteResource.ClientIDs, err = c.GetSiteResourceClients(resID)
				if err != nil {
					return nil, err
				}
				return &siteResource, nil
			}
		}
	}
	return nil, &APIError{ // TODO not a good practice to fake an api error
		StatusCode: 404,
		ApiResponse: ApiResponse{
			Status: 404,
		},
	}
}

func (c *Client) UpdateSiteResource(resID int, res *SiteResource) (*SiteResource, error) {
	path := fmt.Sprintf("/site-resource/%d", resID)
	data, err := c.doRequest("POST", path, res)
	if err != nil {
		return nil, err
	}
	var out SiteResource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteSiteResource(resID int) error {
	path := fmt.Sprintf("/site-resource/%d", resID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

func (c *Client) GetSiteResourceRoles(resID int64) ([]int, error) {
	path := fmt.Sprintf("/site-resource/%d/roles", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Roles []struct {
			RoleID int    `json:"roleId"`
			Name   string `json:"name"`
		} `json:"roles"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(wrapper.Roles)-1)
	for _, r := range wrapper.Roles {
		if r.Name == "Admin" {
			continue // remove implied admin role to avoid conflict
		}
		ids = append(ids, r.RoleID)
	}
	return ids, nil
}

func (c *Client) GetSiteResourceUsers(resID int64) ([]string, error) {
	path := fmt.Sprintf("/site-resource/%d/users", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Users []struct {
			UserID string `json:"userId"`
		} `json:"users"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]string, len(wrapper.Users))
	for i, u := range wrapper.Users {
		ids[i] = u.UserID
	}
	return ids, nil
}

func (c *Client) GetSiteResourceClients(resID int64) ([]int, error) {
	path := fmt.Sprintf("/site-resource/%d/clients", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Clients []struct {
			ClientID int `json:"clientId"`
		} `json:"clients"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	ids := make([]int, len(wrapper.Clients))
	for i, cl := range wrapper.Clients {
		ids[i] = cl.ClientID
	}
	return ids, nil
}

// Resource definitions
type Resource struct {
	ID         int     `json:"resourceId,omitempty"`
	Name       string  `json:"name"`
	Protocol   *string `json:"protocol,omitempty"`
	Http       *bool   `json:"http,omitempty"`
	ProxyPort  *int32  `json:"proxyPort,omitempty"`
	Subdomain  *string `json:"subdomain,omitempty"`
	DomainID   *string `json:"domainId,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty"`
	SSO        *bool   `json:"sso,omitempty"`
	ApplyRules *bool   `json:"applyRules,omitempty"`
}

func (c *Client) CreateResource(orgID string, res Resource) (*Resource, error) {
	res.Enabled = nil
	res.SSO = nil
	res.ApplyRules = nil
	path := fmt.Sprintf("/org/%s/resource", orgID)
	data, err := c.doRequest("PUT", path, res)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetResource(resID int) (*Resource, error) {
	path := fmt.Sprintf("/resource/%d", resID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateResource(resID int, res Resource) (*Resource, error) {
	res.Http = nil
	res.Protocol = nil
	path := fmt.Sprintf("/resource/%d", resID)
	data, err := c.doRequest("POST", path, res)
	if err != nil {
		return nil, err
	}
	var out Resource
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteResource(resID int) error {
	path := fmt.Sprintf("/resource/%d", resID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// Organization definitions
type Organization struct {
	ID            string  `json:"orgId,omitempty"`
	Name          string  `json:"name"`
	Subnet        *string `json:"subnet,omitempty"`
	UtilitySubnet *string `json:"utilitySubnet,omitempty"`
}

func (c *Client) CreateOrganization(org Organization) (*Organization, error) {
	data, err := c.doRequest("PUT", "/org", org)
	if err != nil {
		return nil, err
	}
	var newOrg Organization
	out := struct {
		Org *Organization `json:"org"`
	}{Org: &newOrg}
	err = json.Unmarshal(data, &out)
	return out.Org, err
}

func (c *Client) GetOrganization(orgID string) (*Organization, error) {
	path := fmt.Sprintf("/org/%s", orgID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var newOrg Organization
	out := struct {
		Org *Organization `json:"org"`
	}{Org: &newOrg}
	err = json.Unmarshal(data, &out)
	return out.Org, err
}

func (c *Client) UpdateOrganization(orgID string, org Organization) (*Organization, error) {
	org.Subnet = nil
	org.UtilitySubnet = nil
	path := fmt.Sprintf("/org/%s", orgID)
	data, err := c.doRequest("POST", path, org)
	if err != nil {
		return nil, err
	}
	var udpatedOrg Organization
	out := struct {
		Org *Organization `json:"org"`
	}{Org: &udpatedOrg}
	err = json.Unmarshal(data, &out)
	return out.Org, err
}

func (c *Client) DeleteOrganization(orgID string) error {
	path := fmt.Sprintf("/org/%s", orgID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// Idp definitions
type Idp struct {
	ID                 *int64  `json:"idpId,omitempty"`
	Name               string  `json:"name"`
	ClientID           string  `json:"clientId"`
	ClientSecret       string  `json:"clientSecret"`
	AuthURL            string  `json:"authUrl"`
	TokenURL           string  `json:"tokenUrl"`
	IdentifierPath     *string `json:"identifierPath"`
	EmailPath          *string `json:"emailPath,omitempty"`
	NamePath           *string `json:"namePath,omitempty"`
	Scopes             string  `json:"scopes"`
	AutoProvision      *bool   `json:"autoProvision,omitempty"`
	DefaultRoleMapping *string `json:"defaultRoleMapping,omitempty"`
	DefaultOrgMapping  *string `json:"defaultOrgMapping,omitempty"`
	Tags               *string `json:"tags,omitempty"`
}

func (c *Client) CreateIdp(idp Idp) (*Idp, error) {
	idp.DefaultRoleMapping = nil
	idp.DefaultOrgMapping = nil
	data, err := c.doRequest("PUT", "/idp/oidc", idp)
	if err != nil {
		return nil, err
	}
	out := struct {
		IdpID int64 `json:"idpId"`
	}{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	return c.GetIdp(out.IdpID)
}

func (c *Client) GetIdp(idpID int64) (*Idp, error) {
	path := fmt.Sprintf("/idp/%d", idpID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	out := struct {
		Idp struct {
			ID                 int64   `json:"idpId"`
			Name               string  `json:"name"`
			AutoProvision      bool    `json:"autoProvision"`
			DefaultRoleMapping string  `json:"defaultRoleMapping"`
			DefaultOrgMapping  string  `json:"defaultOrgMapping"`
			Tags               *string `json:"tags"`
		} `json:"idp"`
		IdpOidcConfig struct {
			ClientID       string  `json:"clientId"`
			ClientSecret   string  `json:"clientSecret"`
			AuthURL        string  `json:"authUrl"`
			TokenURL       string  `json:"tokenUrl"`
			IdentifierPath *string `json:"identifierPath"`
			EmailPath      *string `json:"emailPath"`
			NamePath       *string `json:"namePath"`
			Scopes         string  `json:"scopes"`
		} `json:"idpOidcConfig"`
	}{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	return &Idp{
		ID:                 &out.Idp.ID,
		Name:               out.Idp.Name,
		AutoProvision:      &out.Idp.AutoProvision,
		DefaultRoleMapping: &out.Idp.DefaultRoleMapping,
		DefaultOrgMapping:  &out.Idp.DefaultOrgMapping,
		Tags:               out.Idp.Tags,
		ClientID:           out.IdpOidcConfig.ClientID,
		ClientSecret:       out.IdpOidcConfig.ClientSecret,
		AuthURL:            out.IdpOidcConfig.AuthURL,
		TokenURL:           out.IdpOidcConfig.TokenURL,
		IdentifierPath:     out.IdpOidcConfig.IdentifierPath,
		EmailPath:          out.IdpOidcConfig.EmailPath,
		NamePath:           out.IdpOidcConfig.NamePath,
		Scopes:             out.IdpOidcConfig.Scopes,
	}, nil
}

func (c *Client) UpdateIdp(idpID int64, res Idp) (*Idp, error) {
	path := fmt.Sprintf("/idp/%d/oidc", idpID)
	_, err := c.doRequest("POST", path, res)
	if err != nil {
		return nil, err
	}
	return c.GetIdp(idpID)
}

func (c *Client) DeleteIdp(idpID int64) error {
	path := fmt.Sprintf("/idp/%d", idpID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// Target definitions
type Target struct {
	ID                  int64          `json:"targetId,omitempty"`
	ResourceID          *int64         `json:"resourceId,omitempty"`
	SiteID              int64          `json:"siteId"`
	IP                  string         `json:"ip"`
	Port                int32          `json:"port"`
	Method              *string        `json:"method,omitempty"`
	Enabled             *bool          `json:"enabled,omitempty"`
	HCEnabled           *bool          `json:"hcEnabled,omitempty"`
	HCPath              *string        `json:"hcPath,omitempty"`
	HCScheme            *string        `json:"hcScheme,omitempty"`
	HCMode              *string        `json:"hcMode,omitempty"`
	HCHostname          *string        `json:"hcHostname,omitempty"`
	HCPort              *int32         `json:"hcPort,omitempty"`
	HCInterval          *int64         `json:"hcInterval,omitempty"`
	HCUnhealthyInterval *int64         `json:"hcUnhealthyInterval,omitempty"`
	HCTimeout           *int64         `json:"hcTimeout,omitempty"`
	HCHeaders           []TargetHeader `json:"hcHeaders,omitempty"`
	HCFollowRedirects   *bool          `json:"hcFollowRedirects,omitempty"`
	HCMethod            *string        `json:"hcMethod,omitempty"`
	HCStatus            *int64         `json:"hcStatus,omitempty"`
	HCTlsServerName     *string        `json:"hcTlsServerName,omitempty"`
	Path                *string        `json:"path,omitempty"`
	PathMatchType       *string        `json:"pathMatchType,omitempty"`
	RewritePath         *string        `json:"rewritePath,omitempty"`
	RewritePathType     *string        `json:"rewritePathType,omitempty"`
	Priority            *int32         `json:"priority,omitempty"`
}

type TargetHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) CreateTarget(target Target) (*Target, error) {
	path := fmt.Sprintf("/resource/%d/target", *target.ResourceID)
	target.ResourceID = nil
	data, err := c.doRequest("PUT", path, target)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetTarget(targetID int) (*Target, error) {
	path := fmt.Sprintf("/target/%d", targetID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) UpdateTarget(targetID int, target Target) (*Target, error) {
	path := fmt.Sprintf("/target/%d", targetID)
	target.ResourceID = nil
	data, err := c.doRequest("POST", path, target)
	if err != nil {
		return nil, err
	}
	var out Target
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteTarget(targetID int64) error {
	path := fmt.Sprintf("/target/%d", targetID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// Rule Definitions
type Rule struct {
	ID         int64  `json:"ruleId,omitempty"`
	ResourceID *int64 `json:"resourceId,omitempty"`
	Action     string `json:"action"`
	Match      string `json:"match"`
	Value      string `json:"value"`
	Priority   int64  `json:"priority"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

func (c *Client) CreateRule(rule Rule) (*Rule, error) {
	path := fmt.Sprintf("/resource/%d/rule", *rule.ResourceID)
	rule.ResourceID = nil
	data, err := c.doRequest("PUT", path, rule)
	if err != nil {
		return nil, err
	}
	var out Rule
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) GetRule(ruleID int64, resourceID int64) (*Rule, error) {
	path := fmt.Sprintf("/resource/%d/rules", resourceID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Rules []Rule `json:"rules"`
	}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return nil, err
	}
	for _, rule := range out.Rules {
		if rule.ID == ruleID {
			return &rule, nil
		}
	}
	return nil, &APIError{ // TODO not a good practice to fake an api error
		StatusCode: 404,
		ApiResponse: ApiResponse{
			Status: 404,
		},
	}
}

func (c *Client) UpdateRule(ruleID int64, rule Rule) (*Rule, error) {
	path := fmt.Sprintf("/resource/%d/rule/%d", *rule.ResourceID, ruleID)
	rule.ResourceID = nil
	data, err := c.doRequest("POST", path, rule)
	if err != nil {
		return nil, err
	}
	var out Rule
	err = json.Unmarshal(data, &out)
	return &out, err
}

func (c *Client) DeleteRule(ruleID int64, resID int64) error {
	path := fmt.Sprintf("/resource/%d/rule/%d", resID, ruleID)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}
