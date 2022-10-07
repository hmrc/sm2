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
type ProgressWriter struct {
	service       string
	contentLength int
	totalRead     int
	lastMark      int
	renderer      *ProgressRenderer
}

func (pt *ProgressWriter) Write(p []byte) (int, error) {
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
	errors     map[string]error
	updateChan chan Progress
	serviceLen int
	noProgress bool
}

func (pr *ProgressRenderer) init(services []ServiceAndVersion) {

	pr.updateChan = make(chan Progress, 2)
	pr.state = map[string]Progress{}
	pr.errors = map[string]error{}
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

		pendingStart := 0
		maxLines := 20
		if maxLines > len(pr.watchlist) {
			maxLines = len(pr.watchlist)
		}

		for i, service := range pr.watchlist {
			if p, ok := pr.state[service]; ok && p.state == "Pending" {
				pendingStart = i
				break
			}
		}

		drawFrom := 0
		drawTo := maxLines
		if pendingStart > maxLines {
			drawFrom = pendingStart - maxLines
			drawTo = maxLines + drawFrom
		}
		// draw all the stuff
		linesDrawn = 0
		for _, service := range pr.watchlist[drawFrom:drawTo] {
			if p, ok := pr.state[service]; ok {
				fmt.Printf(" %s [%-20s][%3.0f%%] %s\n", crop(pad(p.service, pr.serviceLen), 40), strings.Repeat("=", int(p.percent/5)), p.percent, crop(p.state, 8))
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

func (pr *ProgressRenderer) error(service string, err error) {
	pr.errors[service] = err
}
