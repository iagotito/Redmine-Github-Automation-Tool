package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
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

const RedmineHelpText = `Name
  rgat redmine - Redmine actions

Synopsis
  rgat redmine [-read-yaml <path to yaml file>] [-help]

Description
  Redmine actions

Options
  --read-yaml <path to yaml file>
    Read a sprint info from a YAML file and print on the screen

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

    redmineCmd := flag.NewFlagSet("redmine", flag.ExitOnError)
    redmineHelp := redmineCmd.Bool("help", false, "Help")

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
            return
        }
        setConfig(config{*configUsername, *configPassword})
    case "redmine":
        redmineCmd.Parse(os.Args[2:])
        if *redmineHelp {
            fmt.Println(RedmineHelpText)
            return
        }
        sprint, err := readSprintYaml("sprint1.yaml")
        check(err)
        fmt.Println(sprint)
    default:
        fmt.Println(HelpText)
    }
}

func setConfig(conf config) {
    fmt.Println(conf)
}

type Sprint struct {
    SprintName string `yaml:"name"`
    Tasks []Task `yaml:"tasks"`
}

type Task struct {
    Name string `yaml:"name"`
    Description string `yaml:"description"`
    Subtasks []Task `yaml:"subtasks"`
    Time float32 `yaml:"time"`
}

func readSprintYaml(yamlPath string) (Sprint, error) {
    yfile, err := ioutil.ReadFile(yamlPath)
    if err != nil { return Sprint{}, err }

    var sprint Sprint
    err = yaml.Unmarshal(yfile, &sprint)
    check(err)

    return sprint, nil
}
