/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const PRINT_GAP = 2 // number of spaces between flag and help text or available commands and short description
const BEGIN_GAP = 2 // number of spaces before flag or available commands

type Flag struct {
	name         string
	longName     string
	value        any
	aliases      []string
	ptr          any
	isShort      bool
	isRequire    bool
	isHidden     bool
	usage        string
	introduction string
}

// Usage returns the usage string for the flag.
func (f *Flag) Usage() string {
	if f.usage != "" {
		return f.usage
	}
	aliases := append(f.aliases, f.longName)
	if f.isShort {
		f.usage = fmt.Sprintf("-%s", f.name) + (", --" + strings.Join(aliases, ", --"))
	} else {
		f.usage = ("    --" + strings.Join(aliases, ", --"))
	}
	return f.usage
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func wrapText(text string, width int) []string {
	var lines []string
	runes := []rune(text)
	for len(runes) > width {
		idx := strings.LastIndex(string(runes[:width]), " ")
		if idx == -1 {
			idx = width
		}
		lines = append(lines, string(runes[:idx]))
		runes = runes[idx+1:]
	}

	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}

	return lines
}

func (cmd *Command) strFormat(str string, pos int) string {
	lines := strings.Split(str, "\n")
	if len(lines) != 1 {
		return str
	}
	if str == "" {
		return "\n"
	}
	line_width := max(termSize()-pos, 11)
	lines = wrapText(str, line_width)
	var text string
	for i, line := range lines {
		if i == 0 {
			text += fmt.Sprintf("%s\n", line)
		} else {
			text += fmt.Sprintf("%*s%s\n", pos, " ", line)
		}
	}
	return text
}

func termSize() int {
	w, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// default
		return 80
	}
	return w
}

type Command struct {
	*cobra.Command
	val          reflect.Value
	flags        []*Flag
	flagMap      map[string]*Flag
	maxFlagLen   int
	requireFlag  []*Flag
	optionalFlag []*Flag
	globalFlag   []*Flag
	helpStart    int
	originalPreRunE     func(cmd *cobra.Command, args []string) error
}

func NewCommand(cmd *cobra.Command) *Command {
	t := &Command{
		Command: cmd,
		val:     reflect.ValueOf(cmd.Flags()),
		flags:   []*Flag{},
		flagMap: map[string]*Flag{},
	}
	t.originalPreRunE = cmd.PreRunE
	cmd.PreRunE = t.preRunE

	// Override the flag error function to customize error handling.
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		stdio.Error(err.Error())
		c.SilenceErrors = true
		return err
	})

	cmd.SetUsageFunc(t.PrintUsageFunc)
	cmd.SetHelpFunc(t.PrintHelpFunc)
	return t
}

func newFlag(names []string, ptr, value any, isRequire bool, usage string) *Flag {
	if len(names) == 0 {
		panic("flag name is required")
	}

	flag := &Flag{isRequire: isRequire, value: value, ptr: ptr, introduction: usage}
	length := len(names)
	if length == 1 {
		if len(names[0]) == 1 {
			panic("only short name is not allowed")
		}
		flag.longName = names[0]
	} else {
		sort.Slice(names, func(i, j int) bool {
			return len(names[i]) < len(names[j])
		})
		flag.name = names[0]
		flag.longName = names[length-1]

		if len(flag.name) == 1 {
			flag.isShort = true
			flag.aliases = names[1 : length-1]
		} else {
			flag.name = ""
			flag.aliases = names[:length-1]
		}
	}
	return flag
}

func (c *Command) printFlag(flag *Flag) {
	if flag != nil && !flag.isHidden {
		pFlag := c.Flags().Lookup(flag.longName)
		var help = flag.introduction
		if !isBoolFlag(pFlag) && flag.value != "" && flag.value != 0 {
			help = help + ". Default: " + fmt.Sprint(flag.value)
		}
		stdio.PrintfWithoutNewline("%*s%-*s%*s%s", BEGIN_GAP, " ", c.maxFlagLen, flag.usage, PRINT_GAP, " ", c.strFormat(help, max(c.maxFlagLen+BEGIN_GAP+PRINT_GAP, 11)))
	}
}

// printCategoryFlags prints the flags in the given category.
func (c *Command) printCategoryFlags(flags []*Flag, header string) {
	var printHeader = false
	if len(flags) > 0 {
		for _, flag := range flags {
			if !flag.isHidden && !printHeader {
				stdio.Print(header)
				printHeader = true
			}
			c.printFlag(flag)
		}
	}
}

func sortFlags(flags []*Flag) {
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].longName < flags[j].longName
	})
}

func (c *Command) printFlags() {
	c.initFlags()
	c.initFlagsUsage()
	if c.Flags().SortFlags {
		sortFlags(c.requireFlag)
		sortFlags(c.optionalFlag)
		sortFlags(c.globalFlag)
	}

	c.printCategoryFlags(c.requireFlag, "\nFlags:")
	c.printCategoryFlags(c.optionalFlag, "\nOptional Flags:")
	c.printCategoryFlags(c.globalFlag, "\nGlobal Flags:")
}

func (c *Command) printAliases() {
	if len(c.Aliases) > 0 {
		var aliases []string = []string{c.Name()}
		aliases = append(aliases, c.Aliases...)
		stdio.Print("\nAliases:")
		stdio.Printf("  %s", strings.Join(aliases, ", "))
	}
}

func (c *Command) printAvailableCommands() {
	if c.HasAvailableSubCommands() {
		stdio.Print("\nAvailable Commands:")
		commandLen := 0
		for _, cmd := range c.Commands() {
			if !cmd.Hidden && commandLen < len(cmd.Name()) {
				commandLen = len(cmd.Name())
			}
		}
		for _, cmd := range c.Commands() {
			if !cmd.Hidden {
				stdio.PrintfWithoutNewline("%*s%-*s%*s%s", BEGIN_GAP, " ", commandLen, cmd.Name(), PRINT_GAP, " ", c.strFormat(cmd.Short, max(commandLen+BEGIN_GAP+PRINT_GAP, 11)))
			}
		}
	}
}

func (c *Command) printUseMessage() {
	if c.HasAvailableSubCommands() {
		stdio.Printf("\nUse \"%s [command] --help\" for more information about a command.", GetCmdParentsName(c.Command))
	}
}

func (c *Command) PrintUsageFunc(cmd *cobra.Command) error {
	c.PrintHelpFunc(cmd, nil)
	return nil
}

func (c *Command) PrintHelpFunc(cmd *cobra.Command, args []string) {
	stdio.SetSilenceMode(true)

	// print command short description
	if cmd.Short != "" {
		stdio.Print(cmd.Short + "\n")
	}

	// print usage./
	stdio.Print("Usage:")
	if cmd.HasAvailableSubCommands() {
		stdio.Printf("  %s [command]", GetCmdParentsName(cmd))
	} else if cmd.Flags().HasFlags() {
		stdio.Printf("  %s [flags]", GetCmdParentsName(cmd))
	}

	c.printAliases()
	c.printAvailableCommands()
	c.printFlags()
	c.printUseMessage()
	c.printExample()
}

func (cmd *Command) initFlagsUsage() {
	for _, f := range append(append(cmd.requireFlag, cmd.optionalFlag...), cmd.globalFlag...) {
		usage := f.Usage()
		if len(usage) > cmd.maxFlagLen {
			cmd.maxFlagLen = len(usage)
		}
	}
}

func (cmd *Command) initFlags() {
	cmd.requireFlag = []*Flag{}
	cmd.optionalFlag = []*Flag{}
	for _, flag := range cmd.flags {
		if flag.longName == constant.FLAG_VERBOSE {
			cmd.globalFlag = append(cmd.globalFlag, flag)
		} else if flag.isRequire {
			cmd.requireFlag = append(cmd.requireFlag, flag)
		} else {
			cmd.optionalFlag = append(cmd.optionalFlag, flag)
		}
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		f := cmd.flagMap[flag.Name]
		if f == nil {
			f = &Flag{name: flag.Name, longName: flag.Name, value: flag.DefValue, introduction: flag.Usage}
			if flag.Shorthand != "" {
				f.name = flag.Shorthand
				f.isShort = true
			}
			anno := flag.Annotations[cobra.BashCompOneRequiredFlag]
			if len(anno) > 0 {
				if anno[0] == "true" {
					f.isRequire = true
				}
			}

			if flag.Name == constant.FLAG_HELP {
				cmd.globalFlag = append(cmd.globalFlag, f)
			} else if f.isRequire {
				cmd.requireFlag = append(cmd.requireFlag, f)
			} else {
				cmd.optionalFlag = append(cmd.optionalFlag, f)
			}
		}
		if flag.Hidden {
			f.isHidden = true
		}
	})

}

func (cmd *Command) printExample() {
	if cmd.Command.Example == "" {
		return
	}
	stdio.Print("\nExamples:")
	stdio.Print(cmd.Command.Example)
}

func (cmd *Command) preRunE(ccmd *cobra.Command, args []string) error {
	if err := cmd.flagCheck(); err != nil {
		stdio.Error(err.Error())
		ccmd.SilenceErrors = true
		return err
	}
	if cmd.originalPreRunE != nil {
		if err := cmd.originalPreRunE(cmd.Command, args); err != nil {
			stdio.Error(err.Error())
			cmd.SilenceErrors = true
			return err
		}
	} else if cmd.PreRun != nil {
		cmd.PreRun(ccmd, args)
	}
	return nil
}

func (cmd *Command) flagCheck() (err error) {
	var missingFlagNames = make([]string, 0)

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		f := cmd.flagMap[flag.Name]
		if f == nil || !f.isRequire || flag.Name != f.longName {
			return
		}
		if !flag.Changed && reflect.ValueOf(f.ptr).Elem().IsZero() {
			missingFlagNames = append(missingFlagNames, f.longName)
		}
	})
	if len(missingFlagNames) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missingFlagNames, `", "`))
	}
	return err
}

func (cmd *Command) VarsPs(p any, names []string, value any, usage string, isRequire bool) {
	flag := newFlag(names, p, value, isRequire, usage)
	cmd.flags = append(cmd.flags, flag)
	for _, name := range names {
		if cmd.flagMap[name] != nil {
			panic(fmt.Sprintf("flag %s already exists", name))
		}
		cmd.flagMap[name] = flag
	}

	pType := reflect.TypeOf(value)
	typeNmae := strings.Title(pType.Name())
	funcNmae := fmt.Sprintf("%sVar", typeNmae)

	ptr := reflect.ValueOf(p)
	usageVal := reflect.ValueOf(usage)
	valueVal := reflect.ValueOf(value)

	if flag.isShort {
		method := cmd.val.MethodByName(funcNmae + "P")
		method.Call([]reflect.Value{ptr, reflect.ValueOf(flag.longName), reflect.ValueOf(flag.name), valueVal, usageVal})
	}

	method := cmd.val.MethodByName(funcNmae)
	if !flag.isShort {
		method.Call([]reflect.Value{ptr, reflect.ValueOf(flag.longName), valueVal, usageVal})
	}
	for _, name := range flag.aliases {
		method.Call([]reflect.Value{ptr, reflect.ValueOf(name), valueVal, usageVal})
	}
}

func GetCmdParentsName(cmd *cobra.Command) string {
	var parents []string
	for parent := cmd; parent != nil; parent = parent.Parent() {
		parents = append(parents, parent.Use)
	}
	return strings.Join(ReverseStringSlice(parents), " ")
}

func ReverseStringSlice(slice []string) []string {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

// isBoolFlag checks if a given pflag.Flag is of bool type
func isBoolFlag(flag *pflag.Flag) bool {
	return strings.ToLower(flag.Value.Type()) == "bool"
}
