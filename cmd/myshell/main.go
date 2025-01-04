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

		tokens := strings.Split(command, " ")
		c, err := getCommand(tokens[0])
		if err != nil {
			fmt.Fprintf(os.Stdout, "%s: command not found\n", tokens[0])
			continue
		}

		c.Func(c, tokens[1:])
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
		fmt.Fprintf(os.Stdout, "cd: too many arguments\n")
		return
	}

	err := os.Chdir(args[0])
	if err != nil {
		fmt.Fprintf(os.Stdout, "cd: %s: No such file or directory\n", args[0])
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
