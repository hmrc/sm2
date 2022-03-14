package servicemanager

import (
	"fmt"
	"strings"
)

type Progress struct {
	service string
	percent float32
	state   string
	done    bool
}

// generates progress messages, should be Tee'd from another stream
type ProgressTracker struct {
	service       string
	contentLength int
	totalRead     int
	lastMark      int
	update        chan Progress
}

func (pt *ProgressTracker) Write(p []byte) (int, error) {
	pt.totalRead += len(p)
	pt.lastMark += len(p)

	// send update every 1mb
	if pt.lastMark > (1024 * 1024) {
		pt.lastMark = 0
		percent := (float32(pt.totalRead) / float32(pt.contentLength)) * 100.0
		pt.update <- Progress{service: pt.service, percent: percent, state: "Installing"}
	}
	return len(p), nil
}

type ProgressRenderer struct {
	watchlist  []string
	state      map[string]Progress
	updates    chan Progress
	serviceLen int
}

func (pr *ProgressRenderer) init(services []ServiceAndVersion) {

	pr.state = map[string]Progress{}
	pr.serviceLen = 14

	for _, s := range services {
		pr.watchlist = append(pr.watchlist, s.service)
		pr.state[s.service] = Progress{
			service: s.service,
			state:   "Pending",
		}
		if len(s.service) > pr.serviceLen {
			pr.serviceLen = len(s.service)
		}
	}
}

func (pr *ProgressRenderer) renderLoop(noProgress bool) {

	linesDrawn := 0

	for {
		u := <-pr.updates
		if _, ok := pr.state[u.service]; ok {
			pr.state[u.service] = u
		}

		if !noProgress {
			// clear
			fmt.Print(strings.Repeat("\033[F\033[2K\r", linesDrawn))

			// draw all the stuff
			linesDrawn = 0
			for _, service := range pr.watchlist {
				if p, ok := pr.state[service]; ok {
					fmt.Printf(" %s [%-20s][%3.0f%%] %s\n", pad(p.service, pr.serviceLen), strings.Repeat("=", int(p.percent/5)), p.percent, crop(p.state, 40))
					linesDrawn++
				}
			}
		}
	}
}
