package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const HelpText = `usage: rgat <command> [<args>]

Avaliable commands:

  config        Crate a configuration file

'rgat help' list available subcommands`

const ConfigHelpText = `Name
  rgat config - Configure access information

Synopsis
  rgat config [-username <username>] [-password <password>] [-help]

Description
  Configure the access information.
  Run 'rgat config' without options to open an interactive configuration.

Options
  --username
    Username

  --passowrd
    Password

  --help
    Show help text`

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

type config struct {
    username string
    password string
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println(HelpText)
    }

    configCmd := flag.NewFlagSet("config", flag.ExitOnError)
    configUsername := configCmd.String("username", "", "Username")
    configPassword := configCmd.String("passowrd", "", "Password")
    configHelp := configCmd.Bool("help", false, "Help")

    switch os.Args[1] {

    case "help":
        if len(os.Args) != 2 {
            fmt.Printf("Unknow args for help: %v\n", os.Args[2])
        }
        fmt.Println(HelpText)
    case "config":
        configCmd.Parse(os.Args[2:])
        if *configHelp {
            fmt.Println(ConfigHelpText)
        }
        setConfig(config{*configUsername, *configPassword})
    default:
        fmt.Println(HelpText)
    }
}

func setConfig(conf config) {
    fmt.Println(conf)
}
