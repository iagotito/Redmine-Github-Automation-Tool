package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

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
    redmineGet := redmineCmd.String("get", "", "Test get issues")

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
            return
        }
        if *redmineGet != "" {
            config, err := readConfigYaml(*redmineConfigFile)
            check(err)
            issuesNames := getIssuesSubjectsByPrefix(*redmineGet, config.ProjectUrl, &config)
            for _, issue := range issuesNames {
                fmt.Println(issue)
            }
            return
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
    IssuesPrefix string `yaml:"issues-prefix"`
    StartDate string `yaml:"start-date"`
    DueDate string `yaml:"due-date"`
    Issues []Issue `yaml:"issues"`
}

type Issue struct {
    Subject string `yaml:"subject"`
    Description string `yaml:"description"`
    Subissues []Issue `yaml:"subissues"`
    EstimatedHours float32 `yaml:"estimated-hours"`
    TrackerId int `yaml:"tracker_id"`
    Suffix string `yaml:"suffix"`
    // these fields comes from the sprint struct
    StartDate string
    DueDate string
    // these fields comes from the response after the issue creation
    Num int
    Id int
    ParentId int
}

func isNil(value *interface{}) bool {
    switch v := (*value).(type) {
    case int:
        return *value == 0
    case float32:
        return *value == 0
    case string:
        return *value == ""
    default:
        log.Printf("Unexpected type on isNil: %v", reflect.TypeOf(v))
        return true
    }
}

func (issue *Issue) toJson() ([]byte, error) {
    issueJson := make(map[string]map[string]interface {})
    issueMap := make(map[string]interface {})
    issueJson["issue"] = issueMap

    s := reflect.ValueOf(issue).Elem()
    for i := 0; i < s.NumField(); i++ {
        fieldName := yamlToJsonNames[s.Type().Field(i).Name]
        fieldValue := s.Field(i).Interface()
        if fieldName == "" || isNil(&fieldValue) {
            continue
        }
        issueMap[fieldName] = fieldValue
    }

    issueJsonBytes, err := json.Marshal(issueJson)
    if err != nil {
        return nil, err
    }

    return issueJsonBytes, nil
}

func (issue *Issue) buildSuffix(sprint *Sprint, parentIssue *Issue) string {
    p := fmt.Sprintf("%v%d", sprint.IssuesPrefix, parentIssue.Num)
    suffix := strings.Replace(issue.Suffix, "%p", p, -1)
    return suffix
}

var yamlToJsonNames = map[string]string{
    "Subject": "subject",
    "Description": "description",
    "EstimatedHours": "estimated_hours",
    "StartDate": "start_date",
    "DueDate": "due_date",
    "TrackerId": "tracker_id",
    // TODO: test this with an updated redmine version (not working with demo)
    // see: https://www.redmine.org/issues/18834
    "ParentId": "parent_issue_id",
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

func trimSubjectsPrefix (names []string) []string {
    trimedNames := make([]string, len(names))
    for i := 0; i < len(names); i++ {
        trimedNames[i] = names[i][strings.IndexByte(names[i], ' ')+1:]
    }
    return trimedNames
}

func contains(a []string, s string) bool {
    for _, e := range a {
        if e == s {
            return true
        }
    }
    return false
}

func checkDuplicateIssuesSubjects(issues []Issue) {
    for i := 0; i < len(issues)-1; i++ {
        for j := i+1; j < len(issues); j++ {
            if issues[i].Subject == issues[j].Subject {
                check(errors.New(fmt.Sprintf("Duplicate issues subjects: %v", issues[i].Subject)))
            }
        }
    }
}

func createSprintIssues(sprint Sprint, config Config) {
    checkDuplicateIssuesSubjects(sprint.Issues)

    registeredSubjects := getIssuesSubjectsByPrefix(sprint.IssuesPrefix, config.ProjectUrl, &config)
    registeredSubjectsWithoutPrefix := trimSubjectsPrefix(registeredSubjects)

    totalRegisteredIssues := len(registeredSubjects)

    for _, issue := range sprint.Issues {
        // TODO: check subissues to create new ones if not exists
        if contains(registeredSubjectsWithoutPrefix, issue.Subject) {
            continue
        }
        issue.StartDate = sprint.StartDate
        issue.DueDate = sprint.DueDate
        totalRegisteredIssues++
        // TODO: transform it into an issue method: issue.addPrefix(prefix, num )
        issue.Num = totalRegisteredIssues
        issue.Subject = fmt.Sprintf("%v%d: %v", sprint.IssuesPrefix, issue.Num, issue.Subject)

        postIssue(&issue, config)

        if issue.Subissues != nil {
            for _, subissue := range issue.Subissues {
                totalRegisteredIssues++
                subissue.Suffix = subissue.buildSuffix(&sprint, &issue)
                subissue.Subject = fmt.Sprintf("%v%d: %v %v", sprint.IssuesPrefix, totalRegisteredIssues, subissue.Subject, subissue.Suffix)
                subissue.ParentId = issue.Id
                subissue.DueDate = issue.DueDate
                postIssue(&subissue, config)
            }
        }
    }
}

// TODO: this sould be an Issue method?
func postIssue(issue *Issue, config Config) {
    jsonIssue, err := issue.toJson()
    check(err)

    reqBody := bytes.NewReader(jsonIssue)
    url := fmt.Sprintf("%v/issues.json", config.ProjectUrl)
    req, err := http.NewRequest("POST", url, reqBody)
    check(err)

    req.Header.Add("Content-Type", "application/json")
    // TODO: test this header to set author's id when using a admin account
    // see: https://www.redmine.org/projects/redmine/wiki/Rest_api#User-Impersonation
    //req.Header.Add("X-Redmine-Switch-User", config.Username)
    req.SetBasicAuth(config.Username, config.Password)

    client := http.Client{}
    res, err := client.Do(req)
    check(err)
    if res.StatusCode != 201 {
        check(errors.New(fmt.Sprintf("Status code %d during creation of: \"%s\"", res.StatusCode, issue.Subject)))
    }

    defer res.Body.Close()
    resBody, err := ioutil.ReadAll(res.Body)
    check(err)

    var data map[string]map[string]interface {}
    err = json.Unmarshal(resBody, &data)
    check(err)

    issue.Id = int(data["issue"]["id"].(float64))

    fmt.Printf("Issue %d created: \"%v\"\n", issue.Id, issue.Subject)
}

func getIssuesSubjectsByPrefix(issuesPrefix string, projectUrl string, config *Config) []string {
    req, err := http.NewRequest(
        "GET",
        fmt.Sprintf(
            "%v/issues.json?subject=~%v&limit=99",
            projectUrl,
            issuesPrefix,
        ),
        nil,
    )
    check(err)

    req.Header.Add("Content-Type", "application/json")
    req.SetBasicAuth(config.Username, config.Password)

    client := http.Client{}
    res, err := client.Do(req)
    check(err)
    if res.StatusCode != 200 {
        check(errors.New(fmt.Sprintf("Status code %d during fetch of issues with prefix: \"%v\"", res.StatusCode, issuesPrefix)))
    }

    defer res.Body.Close()
    resBody, err := ioutil.ReadAll(res.Body)
    check(err)

    var data map[string]interface {}
    err = json.Unmarshal(resBody, &data)
    check(err)

    issues := data["issues"].([]interface{})
    issuesNames := make([]string, len(issues))
    for i := 0; i < len(issues); i++ {
        issuesNames[i] = issues[i].(map[string]interface{})["subject"].(string)
    }
    return issuesNames
}
