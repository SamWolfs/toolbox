package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/SamWolfs/go-jsoncanvas"
	"github.com/cli/go-gh/v2"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
}

func (m Metadata) Repository() string {
	return fmt.Sprintf("%s/%s", m.Owner, m.Repo)
}

func main() {
	path := flag.String("path", "path-to-jsoncanvas", "Path to the jsoncanvas file")
	flag.Parse()
	c, err := jsoncanvas.DecodeFile(*path)
	if err != nil {
		fmt.Printf("%s", err)
	}
	metadata := getMetadata(c)
	tasks := c.GetNodesByTag("task")
	for _, task := range tasks {
		// 0: tags, 1: title, 2+: body
		parts := strings.Split(*task.Text, "\n")
		title, _ := strings.CutPrefix(parts[1], "# ")
		body := strings.Join(parts[2:], "\n")

		_, err := getIssue(*metadata, title)

		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		fmt.Printf("Creating issue (%s)\n%s\n", title, body)
		gh.Exec("issue", "create", "--repo", metadata.Repository(), "--title", title, "--body", body)
	}
}

type Issue struct {
	Body  string `json:"body"`
	Title string `json:"title"`
}

func getIssue(metadata Metadata, title string) (*[]Issue, error) {
	issues := new([]Issue)

	issueList, _, err := gh.Exec("issue", "list", "--repo", metadata.Repository(), "-S", title, "--json", "title,body")

	if err != nil {
		return nil, fmt.Errorf("Error while reading issues from GitHub: %s", err)
	}

	if err := json.Unmarshal(issueList.Bytes(), issues); err != nil {
		return nil, fmt.Errorf("can't decode issues list: %w", err)
	}

	if len(*issues) >= 1 {
		return issues, fmt.Errorf("Found one or more issues for title: %s", title)
	}

	return nil, nil
}

func getMetadata(c *jsoncanvas.Canvas) *Metadata {
	meta := &Metadata{}
	metaNodes := c.GetNodesByTag("metadata")

	if len(metaNodes) != 1 {
		return nil
	}

	yamlString := strings.TrimPrefix(*metaNodes[0].Text, "#metadata\n")
	yaml.Unmarshal([]byte(yamlString), meta)

	return meta
}
