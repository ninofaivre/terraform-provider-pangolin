package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPangolinSite_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPangolinSiteConfig("test-site-1", "r35trvxvi6ivchq", "aaoefbvchlouozs442t5x67nbbg80e4a9d6eoexe84cwqbn1", "10.88.1.0", "newt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pangolin_site.test", "name", "test-site-1"),
					resource.TestCheckResourceAttr("pangolin_site.test", "newt_id", "r35trvxvi6ivchq"),
					resource.TestCheckResourceAttr("pangolin_site.test", "address", "10.88.1.0"),
					resource.TestCheckResourceAttr("pangolin_site.test", "type", "newt"),
					resource.TestCheckResourceAttrSet("pangolin_site.test", "id"),
				),
			},
			// Update name and address
			{
				Config: testAccPangolinSiteConfig("test-site-1-updated", "r35trvxvi6ivchq", "aaoefbvchlouozs442t5x67nbbg80e4a9d6eoexe84cwqbn1", "10.88.2.0", "newt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pangolin_site.test", "name", "test-site-1-updated"),
					resource.TestCheckResourceAttr("pangolin_site.test", "address", "10.88.2.0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPangolinSite_NoAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPangolinSiteConfigNoAddress("test-site-noaddr", "r35trvxvi6ivchqb", "aaoefbvchlouozs442t5x67nbbg80e4a9d6eoexe84cwqbn2", "newt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pangolin_site.test", "name", "test-site-noaddr"),
					resource.TestCheckResourceAttr("pangolin_site.test", "type", "newt"),
					resource.TestCheckResourceAttrSet("pangolin_site.test", "id"),
				),
			},
		},
	})
}

func testAccPangolinSiteConfig(name, newtID, secret, address, siteType string) string {
	return fmt.Sprintf(`
provider "pangolin" {
  base_url = %[1]q
  token    = %[2]q
}

resource "pangolin_site" "test" {
  org_id  = %[3]q
  name    = %[4]q
  newt_id = %[5]q
  secret  = %[6]q
  address = %[7]q
  type    = %[8]q
}
`, testURL, testToken, testOrgID, name, newtID, secret, address, siteType)
}

func testAccPangolinSiteConfigNoAddress(name, newtID, secret, siteType string) string {
	return fmt.Sprintf(`
provider "pangolin" {
  base_url = %[1]q
  token    = %[2]q
}

resource "pangolin_site" "test" {
  org_id  = %[3]q
  name    = %[4]q
  newt_id = %[5]q
  secret  = %[6]q
  type    = %[7]q
}
`, testURL, testToken, testOrgID, name, newtID, secret, siteType)
}
