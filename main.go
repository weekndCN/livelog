package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

// line json
type Line struct {
	Number    int    `json:"pos"`  //  Line number
	Message   string `json:"out"`  // line message
	Timestamp int64  `json:"time"` // line generated timestamp
}

func main() {
	cmd := exec.Command("df", "-h")

	// create a pipe from stdout & stderr
	output, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// combina stdout and stderr
	cmd.Stderr = cmd.Stdout
	// create a channel to ensure all output received
	done := make(chan struct{})

	// create a scanner to scan output line by line
	scanner := bufio.NewScanner(output)

	// goroutine read
	go func() {
		i := 0
		for scanner.Scan() {
			var line Line
			line.Number = i
			line.Timestamp = time.Now().Unix()
			line.Message = scanner.Text()
			l, _ := json.Marshal(line)
			fmt.Println(string(l))
			i++
		}
		// unblock the channel when all done
		done <- struct{}{}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	<-done
}
