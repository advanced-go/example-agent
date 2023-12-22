package agent

import (
	"fmt"
	"github.com/advanced-go/core/runtime"
	"github.com/advanced-go/example-domain/activity"
	"github.com/advanced-go/example-domain/slo"
	"github.com/advanced-go/example-domain/timeseries"
	"github.com/advanced-go/example-domain/timeseries/entryv2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var args = new(agentArgs)

func init() {
	args.quit = make(chan struct{}, 1)
}

func Run(d time.Duration) {
	args.run(d)
}

func Stop() {
	args.stop()
}

type agentArgs struct {
	test bool
	ts   []entryv2.Entry
	slo  slo.Entry
	quit chan struct{}
}

func (a *agentArgs) run(d time.Duration) {
	go run(d, a.quit, a)
}

func (a *agentArgs) stop() {
	a.quit <- struct{}{}
	close(a.quit)
}

func (a *agentArgs) getTimeseries() runtime.Status {
	if len(a.ts) > 0 {
		return runtime.NewStatusOK()
	}
	status := runtime.NewStatus(http.StatusInternalServerError)
	a.ts, status = timeseries.GetEntryV2(nil, nil)
	if !status.OK() {
		fmt.Printf("agent: error reading timseries data -> %v\n", status)
	}
	return status
}

func (a *agentArgs) activeSLO() runtime.Status {
	if a.test && len(a.slo.Threshold) > 0 {
		return runtime.NewStatusOK()
	}
	entries, status := slo.GetEntry(nil, nil)
	if !status.OK() {
		fmt.Printf("agent: error reading slo data -> %v\n", status)
		return status
	}
	if len(entries) > 0 {
		a.slo = entries[len(entries)-1]
	}
	return runtime.NewStatusOK()
}

func run(interval time.Duration, quit <-chan struct{}, a *agentArgs) {
	tick := time.Tick(interval)
	var status runtime.Status
	var currentId = ""

	for {
		select {
		case <-tick:
			//fmt.Printf("agent: tick\n")
			status = a.getTimeseries()
			if !status.OK() {
				break
			}
			status = a.activeSLO()
			if !status.OK() {
				break
			}
			if currentId == a.slo.Id {
				fmt.Printf("processing skipped : no SLO changes\n")
				break
			}
			currentId = a.slo.Id
			fmt.Printf("processing slo : %v -> %v\n", a.slo.Id, a.slo.Threshold)
			act := Analyze(a.ts, a.slo)
			if len(act) > 0 {
				_, status = activity.PostEntry[[]activity.Entry](nil, "PUT", nil, act)
				if !status.OK() {
					fmt.Printf("agent: error adding activity -> %v\n", status)
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

func Analyze(ts []entryv2.Entry, slo slo.Entry) []activity.Entry {
	var act []activity.Entry

	ms := durationMS(slo.Threshold)
	for _, e := range ts {
		if e.Duration > ms {
			desc := fmt.Sprintf("duration [%v] is over threshold [%v]", e.Duration, ms)

			act = append(act, activity.Entry{
				ActivityID:   "",
				ActivityType: "trace",
				Agent:        "agent-test",
				AgentUri:     "",
				Assignment:   "",
				Controller:   "controller-test",
				Behavior:     "",
				Description:  desc,
			})
		}
	}
	return act
}
