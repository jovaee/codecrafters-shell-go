package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var builtins = []string{
	"echo",
	"exit",
	"type",
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

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

	dirs := strings.Split(path, ":")
	for _, command := range tokens {
		found := false

		// First check if it's a defined builtin
		if indexOf(command, builtins) != -1 {
			fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", command)
			continue
		}

		// If it's not a builtin then search through PATH
		for _, d := range dirs {

			entries, err := os.ReadDir(d)
			if err != nil {
				continue
			}

			for _, e := range entries {
				if e.Name() == command {
					fmt.Fprintf(os.Stdout, "%s is %s/%s\n", command, d, command)
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stdout, "%s: not found\n", command)
		}
	}
}
