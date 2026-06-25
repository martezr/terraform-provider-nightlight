resource "nightlight_switch" "core" {
  name    = "core-switch"
  site_id = nightlight_site.example.id
  type    = "core"
}
