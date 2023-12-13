# terraform-provider-nats
Terraform provider for [nats](https://nats.io/) messaging system.

Example
```terraform
provider "nats" {}

resource "nats_stream" "orders_stream" {
    name      = "orders"
    subjects  = ["order.*"]
    discard   = "new"
    retention = "interest"
}

resource "nats_consumer" "new_order_consumer" {
   stream_name     = nats_stream.orders_stream.name
   name            = "new_order_consumer" 
   mode            = "pull"
   ack_policy      = "explicit"
   filter_subjects = ["orders.created"]
}

output "orders_stream" {
    value = nats_stream.orders_stream
}

output "new_order_consumer" {
    value = nats_consumer.new_order_consumer
}
```

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:
```shell
go install
# Should be installed as `terraform-nats-provider` binary
which terraform-nats-provider
```

## Using the provider

NOT REGISTERED YET TO REGISTRY.

You can use it locally by:
1. Creating the file `~/.terraformrc` and adding the below to it.
```terraform
provider_installation {

  dev_overrides {
      "registry.terraform.io/AhmadElsagheer/nats" = "/path/to/terraform-nats-provider/binary"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```
2. Adding the below to your terraform project.
```terraform
terraform {
    required_providers {
        nats = {
            source = "registry.terraform.io/AhmadElsagheer/nats"
        }
    }
}
```

Check [examples](./examples) and [docs](./docs) for how to use.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

## TODOs
- [ ] Write unit tests 
- [ ] Add logging
- [ ] Add to terraform public registry
