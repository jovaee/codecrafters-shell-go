package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprint(os.Stdout, "Failed to read command")
			continue
		}

		// Clean up val since it contains the delim char
		command := strings.TrimSpace(input)

		switch command {
		case "exit 0":
			os.Exit(0)
		}

		fmt.Fprintf(os.Stdout, "%s: command not found\n", command)
	}

}
