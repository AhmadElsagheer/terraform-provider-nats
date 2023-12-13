package nats

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test__GetStream(t *testing.T) {
	t.Skip()
	c := makeTestClient()
	info, err := c.GetStream("orders")
	require.NoError(t, err)
	data, err := json.Marshal(info)
	require.NoError(t, err)
	fmt.Println(string(data))
}

func Test__GetConsumer(t *testing.T) {
	t.Skip()
	c := makeTestClient()
	info, err := c.GetConsumer("orders", "new_order_consumer")
	require.NoError(t, err)
	data, err := json.Marshal(info)
	require.NoError(t, err)
	fmt.Println(string(data))
}

func makeTestClient() *client {
	return &client{url: "nats://localhost:4222"}
}
