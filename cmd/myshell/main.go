package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
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

		if err == io.EOF {
			return
		}
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

	var args []string

	n := 0
	for {
		// fmt.Printf("=====\n")
		// fmt.Printf("n=%d\n", n)

		var r byte

		if n >= len(s) {
			break
		}

		switch s[n] {
		case '\'':
			r = '\''
		case '"':
			r = '"'
		default:
			r = 0
		}

		// Enclosing characters in single quotes preserves the literal value of each character within the quotes.
		// ie just take the chars as is
		if s[n] == r {
			k := n
			// fmt.Printf("quote\n")

			for {
				// fmt.Printf("quote inner full s=%s\n", s[k:])
				i := strings.IndexByte(s[k+1:], s[n])
				// fmt.Printf("quote inner i=%d\n", i)
				// fmt.Printf("quote inner k=%d\n", k)

				if i == -1 {
					// Invalid quoting
					return []string{}
				}

				i = k + i + 1

				// If two quotes are next to each other treat it as a continuous string
				if i < len(s)-1 && s[i] == s[i+1] {
					k = i + 2
					continue
				} else {
					k = i
					break
				}
			}

			// fmt.Printf("quote inner part s=%s\n", s[n+1:k])

			v := string(s[n]) + string(s[n])
			args = append(args, strings.ReplaceAll(s[n+1:k], v, ""))
			n = k + 1
			continue
		} else if !unicode.IsSpace(rune(s[n])) {
			i := strings.IndexFunc(s[n:], func(r rune) bool {
				return r == '\'' || r == '"'
			})
			// fmt.Printf("normal i=%d\n", i)

			if i == -1 {
				i = len(s) + 1
			}

			args = append(args, strings.Fields(s[n:i-1])...)
			n = i
			continue
		}

		n += 1
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
