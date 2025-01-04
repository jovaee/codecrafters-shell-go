package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type CommandType string

const (
	BUILTIN  CommandType = "builtin"
	EXTERNAL CommandType = "external"
)

type Command struct {
	Name string
	Type CommandType
	Path string
	Func func(Command, []string)
}

var BuiltinRegister = map[string]Command{}

func main() {

	// Load all builtins.
	registerBuiltins()

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

		// echo world test
		// ["echo", "world test"]
		tokens := strings.SplitN(command, " ", 2)

		cname := tokens[0]

		c, err := getCommand(cname)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s: command not found\n", cname)
			continue
		}

		var args []string
		if len(tokens) > 1 {
			args = parseArgs(tokens[1])
		}
		c.Func(c, args)
	}
}

// Load all custom builtins into the register
func registerBuiltins() {
	BuiltinRegister["echo"] = Command{Name: "echo", Type: BUILTIN, Func: echo}
	BuiltinRegister["exit"] = Command{Name: "exit", Type: BUILTIN, Func: exit}
	BuiltinRegister["type"] = Command{Name: "type", Type: BUILTIN, Func: type_}
	BuiltinRegister["pwd"] = Command{Name: "pwd", Type: BUILTIN, Func: pwd}
	BuiltinRegister["cd"] = Command{Name: "cd", Type: BUILTIN, Func: cd}
}

// Parse
func parseArgs(s string) []string {

	var pairs []byte
	var args []string

	n := 0
	for i, c := range s {
		// If a singlequote is found add to pairs if one isn't on the stack
		// If one is on the stack then create a new argument
		if c == '\'' {
			if len(pairs) > 0 && pairs[len(pairs)-1] == '\'' {
				pairs = pairs[:len(pairs)-1]
				args = append(args, s[n:i])
			} else {
				pairs = append(pairs, '\'')
			}

			n = i + 1 // Skip the starting single quote
			continue
		}

		// If not trying to find a matching pair and an empty space is found
		// create a new argument
		if len(pairs) == 0 && c == ' ' {
			args = append(args, s[n:i])
			n = i + 1
		}
	}

	if len(pairs) != 0 {
		fmt.Fprintf(os.Stdout, "arguments not enclosed")
		return []string{}
	}

	if n < len(s) {
		args = append(args, s[n:])
	}
	return args
}

func getCommand(cname string) (Command, error) {
	c, ok := BuiltinRegister[cname]
	if ok {
		return c, nil
	}

	c, err := loadExternalCommand(cname)
	if err == nil {
		return c, nil
	}

	return Command{}, errors.New("Command not found")
}

func loadExternalCommand(cname string) (Command, error) {
	path := os.Getenv("PATH")
	dirs := strings.Split(path, ":")

	for _, d := range dirs {
		_, err := os.Stat(d + "/" + cname)
		if err != nil {
			continue
		}

		return Command{Name: cname, Type: EXTERNAL, Path: d + "/" + cname, Func: execute}, nil
	}

	return Command{}, errors.New("Command not found in path")
}

// Builtins
func execute(c Command, args []string) {
	cmd := exec.Command(c.Name, args...)
	out, _ := cmd.CombinedOutput()

	fmt.Print(string(out))
}

func exit(c Command, args []string) {
	if len(args) != 1 {
		fmt.Println("exit: incorrect number of arguments")
		return
	}

	code, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stdout, "exit %s: invalid exit code", args[0])
		return
	}

	os.Exit(code)
}

func echo(c Command, args []string) {
	fmt.Println(strings.Join(args, " "))
}

func pwd(c Command, args []string) {
	pwd, _ := os.Getwd()
	fmt.Fprintf(os.Stdout, "%s\n", pwd)
}

func cd(c Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stdout, "cd: incorrect amount of arguments\n")
		return
	}

	home := os.Getenv("HOME")
	to := strings.TrimSpace(args[0])
	if to[0] == '~' {
		to = strings.Replace(to, "~", home, 1)
	}

	err := os.Chdir(to)
	if err != nil {
		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", to)
	}
}

func type_(c Command, args []string) {

	for _, a := range args {

		_, ok := BuiltinRegister[a]
		if ok {
			fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", a)
			continue
		}

		c, err := loadExternalCommand(a)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s: not found\n", a)
			continue
		}

		fmt.Fprintf(os.Stdout, "%s is %s\n", a, c.Path)
	}
}
