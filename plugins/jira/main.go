package main

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/jyggen/pingu"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"
	"regexp"
	"strings"
	"sync"
	"time"
)

type plugin struct {
	*jira.Client
	baseUrl string
}

var commandRegex *regexp.Regexp
var version string

func init() {
	commandRegex = regexp.MustCompile("(?:^|[\\W\\D])!([\\w\\d]+-[\\d]+)")
}

func New(c *viper.Viper) pingu.Plugin {
	transport := jira.BasicAuthTransport{
		Username: c.GetString("jira.username"),
		Password: c.GetString("jira.password"),
	}

	httpClient := transport.Client()
	httpClient.Timeout = c.GetDuration("jira.timeout") * time.Second
	client, _ := jira.NewClient(httpClient, c.GetString("jira.base_url"))

	return pingu.Plugin(&plugin{
		Client: client,
		baseUrl: c.GetString("jira.base_url"),
	})
}

func (pl *plugin) Author() pingu.Author {
	return pingu.Author{
		Email: "jonas@stendahl.me",
		Name:  "Jonas Stendahl",
	}
}

func (pl *plugin) Commands() pingu.Commands {
	return pingu.Commands{
		&pingu.Command{
			Description: "Retrieves one or multiple issues from JIRA.",
			Func: pl.postJiraIssue,
			Trigger: commandRegex,
		},
	}
}

func (pl *plugin) Name() string {
	return "Jira"
}

func (pl *plugin) postJiraIssue(pi *pingu.Pingu, ev *slack.MessageEvent) {
	matches := commandRegex.FindAllStringSubmatch(ev.Text, -1)

	if matches == nil {
		return
	}

	issues := make([]string, len(matches))

	for i, match := range matches {
		issues[i] = match[1]
	}

	invalid := make([]string, 0)
	attachments := make([]slack.Attachment, 0)

	var wg sync.WaitGroup

	wg.Add(len(issues))

	for _, issueId := range issues {
		go (func(issueId string) {
			issue, _, err := pl.Issue.Get(issueId, nil)

			if err != nil {
				invalid = append(invalid, issueId)
				pi.Logger().Error(err)
				wg.Done()
				return
			}

			attachments = append(attachments, createIssueAttachment(issue, pl.baseUrl))
			wg.Done()
		})(issueId)
	}

	wg.Wait()

	if len(attachments) > 0 {
		pi.SendAttachments(attachments, "", ev.Channel)
	}

	numOfInvalid := len(invalid)

	if numOfInvalid > 0 {
		errorMessage := strings.Join([]string{
			strings.Join(invalid[:numOfInvalid-1], ", "),
			invalid[numOfInvalid-1],
		}, " and ")

		if errorMessage[0:5] == " and " {
			errorMessage = errorMessage[5:]
		}

		pi.Reply(ev, fmt.Sprintf("I was unable to retrieve %s.", errorMessage))
	}
}

func (pl *plugin) Tasks() pingu.Tasks {
	return pingu.Tasks{}
}

func (pl *plugin) Version() string {
	return version
}

func createIssueAttachment(issue *jira.Issue, baseUrl string) slack.Attachment {
	body := ""

	if issue.Fields.Assignee != nil {
		body += fmt.Sprintf("_Assigned to_ *%s*", issue.Fields.Assignee.DisplayName)
	}

	numOfComponents := len(issue.Fields.Components)

	if numOfComponents > 0 {
		if body == "" {
			body += "_Affecting_ "
		} else {
			body += " _affecting_ "
		}

		components := make([]string, numOfComponents)

		for i, component := range issue.Fields.Components {
			components[i] = fmt.Sprintf("*%s*", component.Name)
		}

		componentString := strings.Join([]string{
			strings.Join(components[:numOfComponents-1], ", "),
			components[numOfComponents-1],
		}, " _and_ ")

		if componentString[0:7] == " _and_ " {
			componentString = componentString[7:]
		}

		body += componentString
	}

	if body != "" {
		body += "."
	}

	link := "<" + baseUrl + "browse/" + issue.Key + "|" + issue.Key + ">"

	return slack.Attachment{
		AuthorIcon: issue.Fields.Status.IconURL,
		AuthorName: issue.Fields.Status.Name + " " + issue.Fields.Type.Name,
		Color:      getIssueColor(issue.Fields.Status),
		Fallback:   issue.Key + ": " + issue.Fields.Summary,
		Pretext:    "*" + link + "*: " + issue.Fields.Summary,
		MarkdownIn: []string{"pretext", "text"},
		Text:       body,
	}
}

func getIssueColor(status *jira.Status) string {
	switch status.StatusCategory.ColorName {
	case "green":
		return "#14892C"
	case "yellow":
		return "#F6C342"
	case "blue-gray":
		return "#4A6785"
	default:
		return "#4A6785"
	}
}
