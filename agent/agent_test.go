package agent

import "fmt"

func Example_durationMS() {
	s := "99.9/800ms"

	d := durationMS(s)
	fmt.Printf("test: durationMS -> %v\n", d)

	//Output:
	//test: durationMS -> 800

}
