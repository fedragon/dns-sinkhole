package blacklist

import (
	"bufio"
	"fmt"
	"strings"
)

func Parse(scanner *bufio.Scanner) <-chan string {
	out := make(chan string)

	go func(ch chan<- string) {
		defer close(out)

		var started bool
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if !started {
				if strings.ToLower(line) == "# start stevenblack" {
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

				ch <- fields[1]
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("scan error: %v\n", err)
		}
	}(out)

	return out
}
