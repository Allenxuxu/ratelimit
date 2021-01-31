# Go rate limiter

A Golang leaky-bucket & token-bucket rate limit implementation, support automatic dynamic rate limit adjustment.

```go
type RateLimit interface {
	Allow() bool
	Take() time.Time
}
```

`Take` will sleep until you can continue.

`Allow` will be Non-blocking.

### Example

#### leaky bucket

```go
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
```

#### token bucket

```go
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
```

#### dynamic rate limit

```go
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
```