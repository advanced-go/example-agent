package agent

import (
	"fmt"
	"github.com/go-ai-agent/core/runtime"
	"github.com/go-ai-agent/example-domain/slo"
	"github.com/go-ai-agent/example-domain/timeseries"
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
				ms := durationMS(slo)
				for _, e := range ts {
					if e.Duration > ms {
						addActivity()
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

func durationMS(s slo.EntryV1) int {
	return 0
}

func addActivity() {
	//act := activity.EntryV1{}
	//
}
