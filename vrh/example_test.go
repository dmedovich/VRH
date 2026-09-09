package vrh_test

import (
	"fmt"

	"github.com/dmedovich/VRH/vrh"
)

func Example() {
	nodes := []uint64{101, 202, 303, 404}

	ring, err := vrh.New(nodes, vrh.WithBackend(vrh.BackendAuto))
	if err != nil {
		panic(err)
	}

	owner := ring.Owner(12345)
	fmt.Println(owner == 101 || owner == 202 || owner == 303 || owner == 404)
	// Output: true
}
