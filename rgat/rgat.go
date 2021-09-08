package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

const HelpText = `usage: rgat <command> [<args>]

Avaliable commands:

  config        Create a configuration file
  redmine       Redmine actions

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
    Read a sprint info from a YAML file and creates issues on Redmine

  --config-flie <path to config yaml file>
    Config file with acess credentials and project url

  --help
    Show help text`

func check(err error) {
    if err != nil {
        log.Fatal(err)
    }
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
    redmineReadYaml := redmineCmd.String("read-yaml", "", "Read sprint YAML file")
    redmineConfigFile := redmineCmd.String("config-file", "", "Path to config YAML file")

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
        setConfig(Config{"http://project.com", *configUsername, *configPassword})
    case "redmine":
        redmineCmd.Parse(os.Args[2:])
        if *redmineHelp {
            fmt.Println(RedmineHelpText)
            return
        }
        if *redmineReadYaml != "" {
            sprint, err := readSprintYaml(*redmineReadYaml)
            check(err)
            config, err := readConfigYaml(*redmineConfigFile)
            check(err)

            createSprintIssues(sprint, config)
        }
    default:
        fmt.Println(HelpText)
    }
}

func setConfig(conf Config) {
    fmt.Println(conf)
}

type Config struct {
    ProjectUrl string `yaml:"project-url"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type Sprint struct {
    SprintNum string `yaml:"sprint"`
    Issues []Issue `yaml:"issues"`
}

type Issue struct {
    Subject string `yaml:"subject"`
    Description string `yaml:"description"`
    Subissues []Issue `yaml:"subissues"`
    EstimatedHours float32 `yaml:"estimated-hours"`
}

var yamlToJsonNames = map[string]string{
    "Subject": "subject",
    "Description": "description",
    "EstimatedHours": "estimated-hours",
    "Subissues": "sub-issues",
}

func readConfigYaml(yamlPath string) (Config, error) {
    yfile, err := ioutil.ReadFile(yamlPath)
    if err != nil { return Config{}, err }

    var config Config
    err = yaml.Unmarshal(yfile, &config)
    check(err)

    return config, nil
}

func readSprintYaml(yamlPath string) (Sprint, error) {
    yfile, err := ioutil.ReadFile(yamlPath)
    if err != nil { return Sprint{}, err }

    var sprint Sprint
    err = yaml.Unmarshal(yfile, &sprint)
    check(err)

    return sprint, nil
}

func createSprintIssues(sprint Sprint, config Config) {
    //sprintNum := sprint.SprintNum

    fmt.Println(sprint.Issues)
    for i := 0; i < len(sprint.Issues); i++ {
        issueBytes, err := makeIssueJsonBytes(&sprint.Issues[i])
        check(err)

        client := http.Client{}

        body := bytes.NewReader(issueBytes)
        url := fmt.Sprintf("%v/issues.json", config.ProjectUrl)
        req, err := http.NewRequest("POST", url, body)
        check(err)

        req.Header.Add("Content-Type", "application/json")
        req.SetBasicAuth(config.Username, config.Password)

        res, err := client.Do(req)
        check(err)
        defer res.Body.Close()

        resBody, err := io.ReadAll(res.Body)
        check(err)

        fmt.Printf("Status: %d\n", res.StatusCode)
        fmt.Printf("Body: %s\n", string(resBody))
    }
}

// TODO: transform it into a Issue "method"
func makeIssueJsonBytes(issue *Issue) ([]byte, error) {
    issueJson := make(map[string]map[string]interface {})
    issueMap := make(map[string]interface {})
    issueJson["issue"] = issueMap

    s := reflect.ValueOf(issue).Elem()
    for i := 0; i < s.NumField(); i++ {
        fieldName := yamlToJsonNames[s.Type().Field(i).Name]
        fieldValue := s.Field(i).Interface()
        issueMap[fieldName] = fieldValue
    }

    issueJsonBytes, err := json.Marshal(issueJson)
    if err != nil {
        return nil, err
    }
    ioutil.WriteFile("sprint.json", issueJsonBytes, os.ModePerm)

    return issueJsonBytes, nil
}
