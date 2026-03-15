package provider

import (
	"testing"

	"github.com/groteck/terraform-provider-pangolin/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	testOrgID = "test-tf"
	testToken = "f1l1v68jvs2j8ix.34fvctzav5t46kdnchztxz6u5ajfxt5wobs4iulv"
	testURL   = "http://localhost:3003/v1" // Integration API port and prefix
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"pangolin": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func testAccPreCheck(t *testing.T) {
	// Verify API is reachable
	c := client.NewClient(testURL, testToken)
	_, err := c.ListSites(testOrgID)
	if err != nil {
		t.Fatalf("API unreachable or invalid credentials: %v", err)
	}
}

func getTestSiteID(t *testing.T) int {
	c := client.NewClient(testURL, testToken)
	sites, err := c.ListSites(testOrgID)
	if err != nil {
		t.Fatalf("failed to list sites: %v", err)
	}

	if len(sites) > 0 {
		return sites[0].ID
	}

	// Create one if it doesn't exist
	site, err := c.CreateSite(testOrgID, &client.Site{
		Name:   "Test Site",
		NewtID: "test-newt-id-123",
		Secret: "test-secret-123456789012345678901234567890",
		Type:   "newt",
	})
	if err != nil {
		t.Fatalf("failed to create test site: %v", err)
	}

	return site.ID
}
