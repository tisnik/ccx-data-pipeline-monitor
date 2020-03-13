/*
Copyright © 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/RedHatInsights/ccx-data-pipeline-monitor/commands"
	"github.com/RedHatInsights/ccx-data-pipeline-monitor/config"
	"github.com/RedHatInsights/ccx-data-pipeline-monitor/oc"
)

var openShiftConfig config.OpenShiftConfig
var ocLogin string
var colorizer aurora.Aurora

func tryToLogin(ocLogin string) {
	stdout, stderr, err := oc.Login(openShiftConfig.URL, ocLogin)
	fmt.Println(stdout)
	fmt.Println(stderr)
	fmt.Println(err == nil)
	fmt.Println(colorizer.Blue("\nDone"))
}

func login() {
	fmt.Print("login: ")
	p, err := terminal.ReadPassword(0)
	if err != nil {
		fmt.Println(colorizer.Red("not set"))
	} else {
		ocLogin = string(p)
		tryToLogin(ocLogin)
	}
}

type simpleCommand struct {
	prefix  string
	handler func()
}

var simpleCommands = []simpleCommand{
	{"bye", commands.Quit},
	{"exit", commands.Quit},
	{"quit", commands.Quit},
	{"login", login},
}

func executeFixedCommand(t string) {
	// simple commands without parameters
	for _, command := range simpleCommands {
		if strings.HasPrefix(t, command.prefix) {
			command.handler()
			return
		}
	}
	fmt.Println("Command not found")
}

func executor(t string) {
	executeFixedCommand(t)
}

func completer(in prompt.Document) []prompt.Suggest {
	firstWord := []prompt.Suggest{
		{Text: "exit", Description: "quit the application"},
		{Text: "quit", Description: "quit the application"},
		{Text: "bye", Description: "quit the application"},
		{Text: "login", Description: "provide login info"},
	}

	blocks := strings.Split(in.TextBeforeCursor(), " ")

	// don't display compation for empty command
	if in.GetWordBeforeCursor() == "" {
		return nil
	}

	// commands consisting of just one word
	return prompt.FilterHasPrefix(firstWord, blocks[0], true)
}

func main() {

	// read configuration first
	viper.SetConfigName("config_my")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
	openShiftConfig = config.ReadOpenShiftConfig()

	// parse command line arguments and flags
	var colors = flag.Bool("colors", true, "enable or disable colors")
	var useCompleter = flag.Bool("completer", true, "enable or disable command line completer")
	// var askForConfirmation = flag.Bool("confirmation", true, "enable or disable asking for confirmation for selected actions (like delete)")
	flag.Parse()

	colorizer = aurora.NewAurora(*colors)
	commands.SetColorizer(colorizer)

	if *useCompleter {
		p := prompt.New(executor, completer)
		p.Run()
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("> ")
		for scanner.Scan() {
			line := scanner.Text()
			executor(line)
			fmt.Print("> ")
		}
	}
}