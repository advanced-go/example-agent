package agent

import (
	"fmt"
	"github.com/advanced-go/core/runtime"
	"github.com/advanced-go/example-domain/activity"
	"github.com/advanced-go/example-domain/slo"
	"github.com/advanced-go/example-domain/timeseries"
	"strconv"
	"strings"
	"time"
)

var args = agentArgs{}

func init() {
	args.quit = make(chan struct{}, 1)
}

func Run() {
	args.run()
}

func Stop() {
	args.stop()
}

type agentArgs struct {
	ts   []timeseries.EntryV2
	slo  slo.EntryV1
	quit chan struct{}
}

func (a agentArgs) run() {
	go run(time.Second*5, a.quit, a)
}

func (a agentArgs) stop() {
	a.quit <- struct{}{}
	close(a.quit)
}

func (a agentArgs) getTimeseries() *runtime.Status {
	//if len(a.ts) > 0 {
	//	return runtime.NewStatusOK()
	//}
	var status *runtime.Status
	a.ts, status = timeseries.GetEntry[[]timeseries.EntryV2](nil, "")
	if !status.OK() {
		fmt.Printf("agent: error reading timseries data -> %v\n", status)
	}
	return status
}

func (a agentArgs) activeSLO() *runtime.Status {
	//if len(a.slo.Threshold) > 0 {
	//		return runtime.NewStatusOK()
	//	}
	entries, status := slo.GetEntry[[]slo.EntryV1](nil, "")
	if !status.OK() {
		fmt.Printf("agent: error reading slo data -> %v\n", status)
		return status
	}
	if len(entries) > 0 {
		a.slo = entries[len(entries)-1]
	}
	return runtime.NewStatusOK()
}

func run(interval time.Duration, quit <-chan struct{}, a agentArgs) {
	tick := time.Tick(interval)
	var status *runtime.Status

	for {
		select {
		case <-tick:
			fmt.Printf("agent: tick\n")
			status = a.getTimeseries()
			if status.OK() {
				status = a.activeSLO()
				if status.OK() {
					act := Analyze(a.ts, a.slo)
					_, status = activity.PostEntry(nil, "PUT", "", "", act)
					if !status.OK() {
						fmt.Printf("agent: error adding activity -> %v\n", status)
					}
				}
			}
		default:
		}
		select {
		case <-quit:
			//fmt.Printf("bye\n")
			return
		default:
		}
	}
}

func durationMS(threshold string) int {
	if len(threshold) == 0 {
		return -1
	}
	s := strings.Split(threshold, "/")
	if len(s) < 2 {
		return -1
	}
	dur, err := ParseDuration(s[1])
	if err != nil {
		return -1
	}
	return int(dur / time.Millisecond)
}

func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	tokens := strings.Split(s, "ms")
	if len(tokens) == 2 {
		val, err := strconv.Atoi(tokens[0])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Millisecond, nil
	}
	tokens = strings.Split(s, "µs")
	if len(tokens) == 2 {
		val, err := strconv.Atoi(tokens[0])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Microsecond, nil
	}
	tokens = strings.Split(s, "m")
	if len(tokens) == 2 {
		val, err := strconv.Atoi(tokens[0])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Minute, nil
	}
	// Assume seconds
	tokens = strings.Split(s, "s")
	if len(tokens) == 2 {
		s = tokens[0]
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return time.Duration(val) * time.Second, nil
}

func Analyze(ts []timeseries.EntryV2, slo slo.EntryV1) []activity.EntryV1 {
	var act []activity.EntryV1

	ms := durationMS(slo.Threshold)
	for _, e := range ts {
		if e.Duration > ms {
			desc := fmt.Sprintf("duration [%v] is over threshold [%v]", e.Duration, ms)
			act = append(act, activity.EntryV1{
				//CreatedTS:    time.Now().UTC(),
				ActivityID:   "",
				ActivityType: "trace",
				Agent:        "agent-test",
				AgentUri:     "",
				Assignment:   "",
				Controller:   "controller-test",
				Behavior:     "",
				Description:  desc,
			})
			//addActivity("trace", "agent-name", slo.Controller, desc)
		}
	}
	return act
}
