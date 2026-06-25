resource "nightlight_datastore" "nfs" {
  name = "nfs-storage"
  type = "nfs"
  path = "192.168.1.20:/exports/nightlight"
}
