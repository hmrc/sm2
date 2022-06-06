package servicemanager

import (
	"fmt"
	"strings"
)

type Progress struct {
	service string
	percent float32
	state   string
}

// generates progress messages, should be Tee'd from another stream
type ProgressTracker struct {
	service       string
	contentLength int
	totalRead     int
	lastMark      int
	renderer      *ProgressRenderer
}

func (pt *ProgressTracker) Write(p []byte) (int, error) {
	pt.totalRead += len(p)
	pt.lastMark += len(p)

	// send update every 1mb
	if pt.lastMark > (1024 * 1024) {
		pt.lastMark = 0
		percent := (float32(pt.totalRead) / float32(pt.contentLength)) * 100.0
		pt.renderer.update(pt.service, percent, "Installing")
	}
	return len(p), nil
}

type ProgressRenderer struct {
	watchlist  []string
	state      map[string]Progress
	updateChan chan Progress
	serviceLen int
	noProgress bool
}

func (pr *ProgressRenderer) init(services []ServiceAndVersion) {

	pr.updateChan = make(chan Progress, 2)
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

func (pr *ProgressRenderer) renderLoop() {

	linesDrawn := 0

	for {
		u := <-pr.updateChan
		if _, ok := pr.state[u.service]; ok {
			pr.state[u.service] = u
		}

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

func (pr *ProgressRenderer) update(service string, percent float32, state string) {
	if !pr.noProgress {
		pr.updateChan <- Progress{service: service, percent: percent, state: state}
	}
}
