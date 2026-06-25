# Look up an image by name and deploy an instance from it.

data "nightlight_image" "ubuntu" {
  name = "ubuntu-24.04"
}

resource "nightlight_instance" "web" {
  name         = "web-01"
  cpu_cores    = 2
  cpu_sockets  = 1
  memory_mb    = 2048
  datastore_id = "defaultdatastore"
  image_id     = data.nightlight_image.ubuntu.id

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
}
