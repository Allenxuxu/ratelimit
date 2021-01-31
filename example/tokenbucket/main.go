package main

import (
	"fmt"
	"time"

	"github.com/Allenxuxu/ratelimit/tokenbucket"
)

func main() {
	rl := tokenbucket.New(100, 100) // per second
	//rl := tokenbucket.New(100, 100, ratelimit.Per(time.Second)) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		if i > 0 {
			fmt.Println(i, now.Sub(prev).Milliseconds())
		}
		prev = now
	}

	// true
	fmt.Println(rl.Allow())
}
