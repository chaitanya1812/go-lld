package main

// go run cmd/cache-ttl/main.go
import (
	"fmt"
	cache "go-lld/cache-ttl"
	"time"
)

func main() {
	c := cache.NewCacheInterface()
	c.Set("key", "val", 5*time.Second)
	val, ok := c.Get("key")
	fmt.Println("ok : ", ok, "val : ", val)
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		if i == 3 {
			c.Set("key", "val2", 5*time.Second)
		}
		// fetch after delay
		val, ok = c.Get("key")
		fmt.Println("ok : ", ok, "val : ", val)
	}

}
