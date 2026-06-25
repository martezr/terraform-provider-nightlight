resource "nightlight_instance" "example" {
  name         = "web-01"
  description  = "Web server"
  cpu_cores    = 2
  cpu_sockets  = 1
  memory_mb    = 2048
  datastore_id = "defaultdatastore"

  storage_disks = [
    {
      index_number = 0
      boot_order   = 1
      size_gb      = 20
      bus_type     = "virtio"
      datastore_id = "defaultdatastore"
    }
  ]

  network_interfaces = [
    {
      index_number = 0
      boot_order   = 2
      bridge_name  = "nl-shared"
      model        = "virtio"
      connected    = true
    }
  ]

  cdroms = [
    {
      index_number = 0
      boot_order   = 3
      path         = "/opt/nightlight/volumes/defaultdatastore/ubuntu-24.04.iso"
      connected    = true
    }
  ]

  tags = {
    env  = "dev"
    team = "platform"
  }
}

output "instance_ip" {
  value = nightlight_instance.example.primary_ip_address
}
