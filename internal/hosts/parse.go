package hosts

import (
	"bufio"
	"strings"
)

type Result struct {
	Domain string
	Err    error
}

const marker = "# start stevenblack"

// Parse starts parsing a Steven Black hosts file and immediately returns a channel of Results, sending to it as parsing progresses.
func Parse(scanner *bufio.Scanner) <-chan Result {
	out := make(chan Result)

	go func(ch chan<- Result) {
		defer close(out)

		var started bool
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if !started {
				if strings.ToLower(line) == marker {
					started = true
					continue
				}
			}

			if strings.HasPrefix(line, "#") {
				continue
			}

			if started {
				fields := strings.Fields(line)
				if len(fields) < 2 {
					continue
				}

				ch <- Result{Domain: fields[1]}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- Result{Err: err}
		}
	}(out)

	return out
}
