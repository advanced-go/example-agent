package agent

import (
	"fmt"
	"github.com/advanced-go/core/access"
	"github.com/advanced-go/example-domain/slo"
	"github.com/advanced-go/example-domain/timeseries/entryv2"
	"time"
)

var entries = []entryv2.Entry{
	{
		Traffic:        "ingress",
		Duration:       800,
		RequestId:      "request-id-1",
		Uri:            "https://access-log.com/example-domain/timeseries/entry",
		Protocol:       "http",
		Host:           "access-log.com",
		Path:           "/example-domain/timeseries/entry",
		Method:         "GET",
		StatusCode:     200,
		ThresholdFlags: "",
		Threshold:      -1,
	},
	{
		Traffic:        "egress",
		Duration:       1000,
		RequestId:      "request-id-2",
		Uri:            "https://access-log.com/example-domain/timeseries/entry",
		Protocol:       "http",
		Host:           "access-log.com",
		Path:           "/example-domain/timeseries/entry",
		Method:         "PUT",
		StatusCode:     202,
		ThresholdFlags: "",
		Threshold:      -1,
	},
}

func Example_durationMS() {
	s := "99.9/800ms"

	d := durationMS(s)
	fmt.Printf("test: durationMS -> %v\n", d)

	s = "99.9/1200ms"

	d = durationMS(s)
	fmt.Printf("test: durationMS -> %v\n", d)

	//Output:
	//test: durationMS -> 800
	//test: durationMS -> 1200

}

func Example_Analyze() {
	act := Analyze(entries, slo.EntryV1{Threshold: "99.9/600ms"})
	if len(act) > 0 {
		for _, a := range act {
			fmt.Printf("test: Analyze() -> %v\n", a.Description)
		}
	} else {
		fmt.Printf("test: Analyze() -> %v\n", act)
	}

	act = Analyze(entries, slo.EntryV1{Threshold: "99.9/1200ms"})
	if len(act) > 0 {
		for _, a := range act {
			fmt.Printf("test: Analyze() -> %v\n", a.Description)
		}
	} else {
		fmt.Printf("test: Analyze() -> %v\n", act)
	}

	act = Analyze(entries, slo.EntryV1{Threshold: "99.9/801ms"})
	if len(act) > 0 {
		for _, a := range act {
			fmt.Printf("test: Analyze() -> %v\n", a.Description)
		}
	} else {
		fmt.Printf("test: Analyze() -> %v\n", act)
	}
	//Output:
	//test: Analyze() -> duration [800] is over threshold [600]
	//test: Analyze() -> duration [1000] is over threshold [600]
	//test: Analyze() -> []
	//test: Analyze() -> duration [1000] is over threshold [801]

}

func Example_Run() {
	access.EnableTestLogger()
	agent := &agentArgs{
		test: true,
		ts:   entries,
		slo:  slo.EntryV1{Controller: "test-controller", Id: "1234", Threshold: "99.9/700ms"},
		quit: make(chan struct{}, 1),
	}
	agent.run(time.Millisecond * 500)
	time.Sleep(time.Millisecond * 1500)
	agent.stop()

	agent2 := &agentArgs{
		test: true,
		ts:   entries,
		slo:  slo.EntryV1{Controller: "test-controller", Id: "5678", Threshold: "99.9/900ms"},
		quit: make(chan struct{}, 1),
	}
	agent2.run(time.Millisecond * 750)
	time.Sleep(time.Millisecond * 1500)
	agent2.stop()

	agent3 := &agentArgs{
		test: false,
		ts:   entries,
		slo:  slo.EntryV1{Controller: "test-controller", Id: "9012", Threshold: "99.9/1200ms"},
		quit: make(chan struct{}, 1),
	}
	agent3.run(time.Millisecond * 1000)
	time.Sleep(time.Millisecond * 1500)
	agent3.stop()

	time.Sleep(time.Millisecond * 1500)
	//fmt.Printf("\n")

	//Output:
	//{ "activity": "trace" "agent": "agent-test"  "controller": "controller-test"  "message": "duration [800] is over threshold [700]"  }
	//{ "activity": "trace" "agent": "agent-test"  "controller": "controller-test"  "message": "duration [1000] is over threshold [700]"  }
	//{ "activity": "trace" "agent": "agent-test"  "controller": "controller-test"  "message": "duration [1000] is over threshold [900]"  }

}
