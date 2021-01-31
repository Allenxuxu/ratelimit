package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Allenxuxu/ratelimit"

	"github.com/Allenxuxu/ratelimit/leakybucket"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	cpu := make(chan float64)
	go mockCPU(ctx, cpu)

	exceptCPU := float64(30) // 30%
	minLimit := int64(10)
	maxLimit := int64(1000)
	rl := leakybucket.New(10, ratelimit.DynamicLimit(ctx, exceptCPU, cpu, minLimit, maxLimit)) // per second

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

func mockCPU(ctx context.Context, c chan float64) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("exit")
			return
		case c <- float64(rand.Intn(100)):
			time.Sleep(time.Second)
		}
	}
}
