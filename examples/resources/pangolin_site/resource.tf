resource "pangolin_site" "example" {
  org_id  = "your-org-id"
  name    = "example-site"
  newt_id = "r35trvxvi6ivchq"
  secret  = var.site_secret
  address = "10.88.1.0"
  type    = "newt"
}
