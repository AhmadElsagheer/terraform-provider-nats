resource "nats_stream" "example" {
    name = "orders"
    subjects = ["order.*"]
    discard = "new"
    retention = "interest"
}
