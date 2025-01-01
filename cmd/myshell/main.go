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

		if command == "" {
			continue
		}

		if command == "exit 0" {
			exit()
		}

		tokens := strings.Split(command, " ")
		switch tokens[0] {
		case "echo":
			echo(tokens[1:])
			continue
		case "type":
			type_(tokens[1:])
			continue
		default:
			fmt.Fprintf(os.Stdout, "%s: command not found\n", command)
		}
	}
}

// Commands
func exit() {
	os.Exit(0)
}

func echo(tokens []string) {
	fmt.Println(strings.Join(tokens, " "))
}

func type_(tokens []string) {
	path := os.Getenv("PATH")

	exes := strings.Split(path, ":")
	for _, token := range tokens {
		found := false

		for _, e := range exes {
			if strings.Contains(e, token) {
				fmt.Fprintf(os.Stdout, "%s is %s\n", token, e)
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stdout, "%s: not found\n", token)
		}
	}
}
