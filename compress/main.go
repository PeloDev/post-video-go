package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func main() {
	// 1. Subscribe to "valid-upload"
	subscriber := redisClient.Subscribe(ctx, "valid-upload")
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		// 2. Not sure... do videos need to be compressed for the purposes of this system? TODO: figure this out
		fmt.Println(msg.Payload)
	}

}
