package main

import (
	"fmt"
	"time"

	"github.com/Allenxuxu/ratelimit/leakybucket"
)

func main() {
	rl := leakybucket.New(10) // per second
	//rl := leakybucket.New(10, ratelimit.Per(time.Second)) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		if i > 0 {
			fmt.Println(i, now.Sub(prev).Milliseconds())
		}
		prev = now
	}

	// false
	fmt.Println(rl.Allow())
}
