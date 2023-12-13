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
