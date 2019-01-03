/*
 * btrfscue version 0.5
 * Copyright (c)2011-2019 Christian Blichmann
 *
 * Command-line sub-commands
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package subcommand // import "blichmann.eu/code/btrfscue/subcommand"

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Command interface {
	DefineFlags(fs *flag.FlagSet)
	Run(arguments []string)
}

type subCommand struct {
	desc string
	cmd  Command
	hide bool
}

type CommandSet struct {
	commands      map[string]*subCommand
	parsedCommand Command
	errorHandling flag.ErrorHandling
	globalFlags   *flag.FlagSet
	args          []string
	output        io.Writer // nil means stderr, use out() internally
}

func NewCommandSet(errorHandling flag.ErrorHandling) *CommandSet {
	return &CommandSet{
		commands:      make(map[string]*subCommand),
		errorHandling: errorHandling,
	}
}

func (c *CommandSet) out() io.Writer {
	if c.output == nil {
		return os.Stderr
	}
	return c.output
}

// SetOutput sets the destination for usage and error messages. If output is
// nil, messages go to standard error.
func (c *CommandSet) SetOutput(output io.Writer) {
	c.output = output
}

func (c *CommandSet) Register(name, desc string, cmd Command) Command {
	c.commands[name] = &subCommand{desc: desc, cmd: cmd}
	return cmd
}

func (c *CommandSet) RegisterHidden(name string, cmd Command) Command {
	c.commands[name] = &subCommand{hide: true, cmd: cmd}
	return cmd
}

func (c *CommandSet) SetGlobalFlags(fs *flag.FlagSet) {
	c.globalFlags = fs
}

func (c *CommandSet) Parse(arguments []string) error {
	var cmdName string
	cmdPos := -1
	// Assume first non-flag is the command's name
	for i, arg := range arguments {
		if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			continue
		}
		cmdPos = i
		cmdName = arg
		break
	}

	// No command given, nothing more to do
	if cmdPos < 0 {
		return nil
	}

	var err error = nil
	if parsed, ok := c.commands[cmdName]; !ok {
		err = fmt.Errorf("'%s' is not a valid command", cmdName)
		fmt.Fprintln(c.out(), err)
	} else {
		fs := flag.NewFlagSet(cmdName, c.errorHandling)
		fs.SetOutput(c.output)
		c.parsedCommand = parsed.cmd
		if c.globalFlags != nil {
			c.globalFlags.VisitAll(func(f *flag.Flag) {
				fs.Var(f.Value, f.Name, f.Usage)
			})
		}
		c.parsedCommand.DefineFlags(fs)
		err = fs.Parse(arguments[cmdPos+1:])
		c.args = fs.Args()
	}
	if err != nil {
		switch c.errorHandling {
		case flag.ContinueOnError:
		case flag.ExitOnError:
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		default:
			panic("unexpected error handling")
		}
	}
	return err
}

func (c *CommandSet) Args() []string {
	return c.args
}

func (c *CommandSet) Run(arguments []string) {
	c.parsedCommand.Run(arguments)
}

func (c *CommandSet) VisitAll(fn func(name, desc string, cmd Command)) {
	names := make([]string, 0, len(c.commands))
	for k, _ := range c.commands {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, n := range names {
		sub := c.commands[n]
		if !sub.hide {
			fn(n, sub.desc, sub.cmd)
		}
	}
}

func Register(name, description string, cmd Command) Command {
	return Commands.Register(name, description, cmd)
}

func Parse(arguments []string) {
	Commands.Parse(arguments)
}

func Run() {
	Commands.Run(Commands.Args())
}

func VisitAll(fn func(name, desc string, cmd Command)) {
	Commands.VisitAll(fn)
}

var Commands = NewCommandSet(flag.ExitOnError)
