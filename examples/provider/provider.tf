terraform {
  required_providers {
    nightlight = {
      source = "martezr/nightlight"
    }
  }
}

provider "nightlight" {
  endpoint = "http://192.168.1.10"
  username = "root"
  password = "nightlight"
}
