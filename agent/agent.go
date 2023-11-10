package agent

import (
	"fmt"
	"github.com/go-ai-agent/core/runtime"
	"github.com/go-ai-agent/example-domain/activity"
	"github.com/go-ai-agent/example-domain/slo"
	"github.com/go-ai-agent/example-domain/timeseries"
	"strconv"
	"strings"
	"time"
)

func Run() {
	//quit <-chan struct{}, status chan *runtime.Status
}

func run(interval time.Duration, quit <-chan struct{}) {
	tick := time.Tick(interval)
	ts, stat := timeseries.Get[[]timeseries.EntryV1](nil, "")
	if !stat.OK() {
		fmt.Printf("error reading timseries data -> %v\n", stat)
		return
	}

	for {
		select {
		case <-tick:
			slo, status := activeSLO()
			if !status.OK() {
				fmt.Printf("error reading active SLO -> %v\n", status)
			} else {
				ms := durationMS(slo.Threshold)
				for _, e := range ts {
					if e.Duration > ms {
						desc := fmt.Sprintf("duration [%v] is over threshold [%v]", e.Duration, ms)
						addActivity("trace", "agent-name", slo.Controller, desc)
					}
				}
			}
		default:
		}
		select {
		case <-quit:
			fmt.Printf("bye\n")
			return
		default:
		}
	}
}

func activeSLO() (slo.EntryV1, *runtime.Status) {
	entries, status := slo.Get[[]slo.EntryV1](nil, "")
	if !status.OK() {
		return slo.EntryV1{}, status
	}
	return entries[len(entries)-1], runtime.NewStatusOK()
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

func addActivity(act, agent, controller, desc string) *runtime.Status {
	e := activity.EntryV1{ActivityType: act, Agent: agent, Controller: controller, Description: desc}
	var entries []activity.EntryV1

	entries = append(entries, e)
	_, status := activity.Do[[]activity.EntryV1](nil, "PUT", "", "", entries)
	return status
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

func Example_Analyze(ts []timeseries.EntryV1) {
	/*
		ms := durationMS(slo.Threshold)
		for _, e := range ts {
			if e.Duration > ms {
				desc := fmt.Sprintf("duration [%v] is over threshold [%v]", e.Duration, ms)
				addActivity("trace", "agent-name", slo.Controller, desc)
			}
		}

	*/
}
