package main

import (
	"snowflake-go"
	"fmt"
)

func main() {
	r := snowflake.NewSnowFlake(0, 0)
	var i int64
	for i = 0; i <= 10000; i++ {
		id := r.NextID()
		fmt.Printf("%d\n", id)
		fmt.Printf("%064b\n", id)
		ids := fmt.Sprintf("%064b", id)
		fmt.Println(ids, len(ids))
	}
}