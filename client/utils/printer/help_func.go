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

package printer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/oceanbase/obshell/client/constant"
	"github.com/oceanbase/obshell/client/lib/http"
	"github.com/oceanbase/obshell/client/lib/stdio"
)

const (
	FLAG_PASSWORD       = "rootpassword"
	FLAG_PASSWORD_ALIAS = "rp"
	FLAG_RS_LIST        = "rs_list"
	FLAG_RS_LIST_ALIAS  = "rs"

	// ALIAS
	ANNOTATIONS_ALIAS = "alias"
)

var customFlag = [][2]string{{"rootpassword", "rp"}}

func PrintHelpFunc(cmd *cobra.Command, requiredFlags []string) {
	stdio.SetSilenceMode(true)

	// get unhidden flags
	cmd.Flags().SortFlags = false
	unhiddenFlags := getUnhiddenFlag(cmd.Flags())

	// print command short description
	stdio.Print(cmd.Short)

	// print usage
	stdio.Print("\nUsage:")
	stdio.Printf("  %s [flags]\n", GetCmdParentsName(cmd))

	// Determine maximum length of flag names to align output
	maxFlagLen := getMaxFlagLen(unhiddenFlags)

	// print required flags
	printFlagSet(unhiddenFlags, requiredFlags, maxFlagLen, true)

	// print optional flags
	printFlagSet(unhiddenFlags, requiredFlags, maxFlagLen, false)

	// print global flags if the command has any
	printGlobalFlagSet(unhiddenFlags, maxFlagLen)
	// print examples
	stdio.Print("Examples:")
	stdio.Print(cmd.Example)
}

func PrintUsageFunc(cmd *cobra.Command) {
	stdio.SetSilenceMode(true)

	// print usage
	stdio.Print("Usage:")
	stdio.Printf("  %s [flags]\n", GetCmdParentsName(cmd))

	// Determine maximum length of flag names to align output
	maxFlagLen := getMaxFlagLen(cmd.Flags())
	// get requiredFlags
	requiredFlags := getRequiredFlags(cmd)

	// print required flags
	printFlagSet(cmd.Flags(), requiredFlags, maxFlagLen, true)

	// print optional flags
	printFlagSet(cmd.Flags(), requiredFlags, maxFlagLen, false)

	// print global flags if the command has any
	printGlobalFlagSet(cmd.Flags(), maxFlagLen)
	// print examples
	stdio.Print("Examples:")
	stdio.Print(cmd.Example)
}

func getRequiredFlags(cmd *cobra.Command) []string {
	requiredFlags := []string{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		anno := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if anno != nil && len(anno) > 0 {
			if anno[0] == "true" {
				requiredFlags = append(requiredFlags, flag.Name)
			}
		}
	})
	return requiredFlags
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

func printGlobalFlagSet(flagSet *pflag.FlagSet, maxFlagLen int) {
	helpFlag := flagSet.Lookup(constant.FLAG_HELP)
	verboseFlag := flagSet.Lookup(constant.FLAG_VERBOSE)
	if helpFlag == nil && verboseFlag == nil {
		return
	}

	stdio.Print("Global Flags:")
	if helpFlag != nil {
		printOneFlag(helpFlag, maxFlagLen)
	}
	if verboseFlag != nil {
		printOneFlag(verboseFlag, maxFlagLen)
	}
	stdio.Print("")
}

func getOptionalFlagSize(flagSet *pflag.FlagSet) int {
	flagSize := flagSetSize(flagSet)
	helpFlag := flagSet.Lookup(constant.FLAG_HELP)
	verboseFlag := flagSet.Lookup(constant.FLAG_VERBOSE)
	if helpFlag != nil {
		flagSize--
	}
	if verboseFlag != nil {
		flagSize--
	}
	return flagSize
}

func getUnhiddenFlag(flagSet *pflag.FlagSet) *pflag.FlagSet {
	unhiddenFlagSet := pflag.NewFlagSet("unhidden", pflag.ContinueOnError)
	flagSet.VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden {
			unhiddenFlagSet.AddFlag(flag)
		}
	})
	return unhiddenFlagSet
}

func printFlagSet(flagSet *pflag.FlagSet, requiredFlags []string, maxFlagLen int, isRequired bool) {
	if isRequired && len(requiredFlags) > 0 {
		stdio.Print("Flags:")
	} else if !isRequired && getOptionalFlagSize(flagSet) > len(requiredFlags) {
		stdio.Print("Optional Flags:")
	} else {
		return
	}

	flagSet.VisitAll(func(flag *pflag.Flag) {
		if http.ContainsString(requiredFlags, flag.Name) != isRequired {
			return
		}

		if flag.Name == constant.FLAG_HELP || flag.Name == constant.FLAG_VERBOSE {
			return
		}

		printOneFlag(flag, maxFlagLen)
	})
	stdio.Print("")
}

func printOneFlag(flag *pflag.Flag, maxFlagLen int) {
	if flag.Hidden {
		return
	}
	// Create a string with the flag name and shorthand
	var flagUsage string
	if len(flag.Shorthand) > 0 {
		flagUsage = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
	} else {
		if flag.Annotations[ANNOTATIONS_ALIAS] != nil {
			flagUsage = fmt.Sprintf("--%s, --%s", flag.Annotations[ANNOTATIONS_ALIAS][0], flag.Name)
		} else {
			flagUsage = fmt.Sprintf("    --%s", flag.Name)
		}
	}

	// Print the flag usage, aligning the output
	usage := flag.Usage
	if !isBoolFlag(flag) && flag.DefValue != "" {
		// Print the default value if it is set
		usage = fmt.Sprintf("%s. Default: %s", flag.Usage, flag.DefValue)
	}
	stdio.Printf("  %-*s %s", maxFlagLen, flagUsage, usage)
}

// isBoolFlag checks if a given pflag.Flag is of bool type
func isBoolFlag(flag *pflag.Flag) bool {
	return strings.ToLower(flag.Value.Type()) == "bool"
}

func getMaxFlagLen(flagSet *pflag.FlagSet) int {
	maxLen := 0
	flagSet.VisitAll(func(flag *pflag.Flag) {
		// Compute the printed length of the flag
		length := len(flag.Name) + 7 // add 4 for the '-, --' prefix
		if length > maxLen {
			maxLen = length
		}
	})
	return maxLen
}

func flagSetSize(flagSet *pflag.FlagSet) int {
	count := 0
	flagSet.VisitAll(func(flag *pflag.Flag) {
		count++
	})
	return count
}
