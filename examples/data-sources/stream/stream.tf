terraform {
    required_providers {
        nats = {
            source = "registry.terraform.io/AhmadElsagheer/nats"
        }
    }
}

provider "nats" {}

data "nats_stream" "orders_stream" {
    name = "orders"
}

output "orders_stream" {
    value = data.nats_stream.orders_stream
}
