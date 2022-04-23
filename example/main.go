package main

import (
	"time"
	"fmt"

	"github.com/treeforest/snowflake"
)

func main() {
	sf, err := snowflake.NewSnowflake(1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("id: %d\n", sf.Generate())

	since := time.Now()
	for i := 0; i < 100000; i++ {
		_ = sf.Generate()
	}
	sub := time.Now().Sub(since)
	fmt.Printf("average used: %dns\n", sub.Nanoseconds()/100000)
}