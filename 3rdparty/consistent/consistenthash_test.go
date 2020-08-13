package consistent_test

import (
	"fmt"
	"testing"

	"github.com/overtalk/bgo/3rdparty/consistent"
)

func TestConsistent(t *testing.T) {
	chash := consistent.New()

	fmt.Println(chash.Get("user1"))

	chash.Add("ip1")
	chash.Add("ip2")

	res := map[string]int{}

	for i := 0; i < 10000; i++ {
		ip, _ := chash.Get(fmt.Sprintf("ip - %d", i))
		res[ip] = res[ip] + 1
	}

	fmt.Println(res)
}
