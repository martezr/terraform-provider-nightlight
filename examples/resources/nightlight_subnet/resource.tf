resource "nightlight_subnet" "example" {
  name         = "dev-subnet"
  cidr_block   = "10.10.0.0/24"
  gateway      = "10.10.0.1"
  dhcp_server  = true
  ip_pool_range = "10.10.0.100-10.10.0.200"
  dns_servers  = ["1.1.1.1", "8.8.8.8"]
  bridge_name  = "nl-shared"
}
