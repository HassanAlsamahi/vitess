/*
Copyright 2021 The Vitess Authors.

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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
)

type (
	label struct {
		Name string `json:"name"`
	}

	labels []label

	author struct {
		Login string `json:"login"`
	}

	prInfo struct {
		Labels labels `json:"labels"`
		Number int    `json:"number"`
		Title  string `json:"title"`
		Author author `json:"author"`
	}

	prsByComponent = map[string][]prInfo

	prsByType = map[string]prsByComponent

	sortedPRComponent struct {
		Name    string
		PrInfos []prInfo
	}

	sortedPRType struct {
		Name       string
		Components []sortedPRComponent
	}

	knownIssue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}

	releaseNote struct {
		Version       string
		Announcement  string
		KnownIssues   string
		AddDetails    string
		ChangeLog     string
		ChangeMetrics string
	}
)

const (
	markdownTemplate = `# Release of Vitess {{.Version}}

{{- if or .Announcement .AddDetails }}
## Announcement
{{ .Announcement }}
{{- end }}

{{- if .AddDetails }}
> TODO: please detail these pull requests.
{{ .AddDetails }}
{{- end }}

{{- if and (or .Announcement .AddDetails) (or .KnownIssues .ChangeLog) }}
------------
{{- end }}

{{- if .KnownIssues }}
## Known Issues
{{ .KnownIssues }}

{{- if .ChangeLog }}
------------
{{- end }}
{{- end }}

{{- if .ChangeLog }}
## Changelog
{{ .ChangeLog }}
{{ .ChangeMetrics }}
{{- end }}
`

	markdownTemplatePR = `
{{- range $type := . }}
### {{ $type.Name }}
{{- range $component := $type.Components }} 
#### {{ $component.Name }}
{{- range $prInfo := $component.PrInfos }}
 * {{ $prInfo.Title }} #{{ $prInfo.Number }}
{{- end }}
{{- end }}
{{- end }}
`

	markdownTemplateKnownIssues = `
{{- range $issue := . }}
 * {{ $issue.Title }} #{{ $issue.Number }} 
{{- end }}
`

	prefixType        = "Type: "
	prefixComponent   = "Component: "
	numberOfThreads   = 10
	lengthOfSingleSHA = 40
)

func (l labels) needsToList() bool {
	for _, label := range l {
		if label.Name == "release notes" {
			return true
		}
	}
	return false
}

func (l labels) needsDetails() bool {
	for _, label := range l {
		if label.Name == "release notes (needs details)" {
			return true
		}
	}
	return false
}

func (rn *releaseNote) generate(writeTo io.Writer) error {
	t := template.Must(template.New("release_notes").Parse(markdownTemplate))
	err := t.ExecuteTemplate(writeTo, "release_notes", rn)
	if err != nil {
		return err
	}
	return nil
}

func loadKnownIssues(release string) ([]knownIssue, error) {
	idx := strings.Index(release, ".")
	if idx > -1 {
		release = release[:idx]
	}
	label := fmt.Sprintf("Known issue: %s", release)
	out, err := execCmd("gh", "issue", "list", "--repo", "vitessio/vitess", "--label", label, "--json", "title,number")
	if err != nil {
		return nil, err
	}
	var knownIssues []knownIssue
	err = json.Unmarshal(out, &knownIssues)
	if err != nil {
		return nil, err
	}
	return knownIssues, nil
}

func loadMergedPRs(from, to string) (prs []string, authors []string, commitCount int, err error) {
	// load the git log with "author \t title \t parents"
	out, err := execCmd("git", "log", `--pretty=format:%ae%x09%s%x09%P%x09%h`, fmt.Sprintf("%s..%s", from, to))

	if err != nil {
		return
	}

	return parseGitLog(string(out))
}

func parseGitLog(s string) (prs []string, authorCommits []string, commitCount int, err error) {
	rx := regexp.MustCompile(`(.+)\t(.+)\t(.+)\t(.+)`)
	mergePR := regexp.MustCompile(`Merge pull request #(\d+)`)
	squashPR := regexp.MustCompile(`\(#(\d+)\)`)
	authMap := map[string]string{} // here we will store email <-> gh user mappings
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		lineInfo := rx.FindStringSubmatch(line)
		if len(lineInfo) != 5 {
			log.Fatalf("failed to parse the output from git log: %s", line)
		}
		authorEmail := lineInfo[1]
		title := lineInfo[2]
		parents := lineInfo[3]
		sha := lineInfo[4]
		merged := mergePR.FindStringSubmatch(title)
		if len(merged) == 2 {
			// this is a merged PR. remember the PR #
			prs = append(prs, merged[1])
			continue
		}

		if len(parents) <= lengthOfSingleSHA {
			// we have a single parent, and the commit counts
			commitCount++
			if _, exists := authMap[authorEmail]; !exists {
				authMap[authorEmail] = sha
			}
		}

		squashed := squashPR.FindStringSubmatch(title)
		if len(squashed) == 2 {
			// this is a merged PR. remember the PR #
			prs = append(prs, squashed[1])
			continue
		}
	}

	for _, author := range authMap {
		authorCommits = append(authorCommits, author)
	}

	sort.Strings(prs)
	sort.Strings(authorCommits) // not really needed, but makes testing easier

	return
}

func execCmd(name string, arg ...string) ([]byte, error) {
	out, err := exec.Command(name, arg...).Output()
	if err != nil {
		execErr, ok := err.(*exec.ExitError)
		if ok {
			return nil, fmt.Errorf("%s:\nstderr: %s\nstdout: %s", err.Error(), execErr.Stderr, out)
		}
		if strings.Contains(err.Error(), " executable file not found in") {
			return nil, fmt.Errorf("the command `gh` seems to be missing. Please install it from https://github.com/cli/cli")
		}
		return nil, err
	}
	return out, nil
}

func loadPRInfo(pr string) (prInfo, error) {
	out, err := execCmd("gh", "pr", "view", pr, "--json", "title,number,labels,author")
	if err != nil {
		return prInfo{}, err
	}
	var prInfo prInfo
	err = json.Unmarshal(out, &prInfo)
	return prInfo, err
}

func loadAuthorInfo(sha string) (string, error) {
	out, err := execCmd("gh", "api", "/repos/vitessio/vitess/commits/"+sha)
	if err != nil {
		return "", err
	}
	var prInfo prInfo
	err = json.Unmarshal(out, &prInfo)
	if err != nil {
		return "", err
	}
	return prInfo.Author.Login, nil
}

type req struct {
	isPR bool
	key  string
}

func loadAllPRs(prs, authorCommits []string) ([]prInfo, []prInfo, []string, error) {
	errChan := make(chan error)
	wgDone := make(chan bool)
	prChan := make(chan req, len(prs)+len(authorCommits))
	// fill the work queue
	for _, s := range prs {
		prChan <- req{isPR: true, key: s}
	}
	for _, s := range authorCommits {
		prChan <- req{isPR: false, key: s}
	}
	close(prChan)

	var prInfos, prNeedsDetails []prInfo
	var authors []string
	fmt.Printf("Found %d merged PRs. Loading PR info", len(prs))
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	shouldLoad := func(in string) bool {
		if in == "" {
			return false
		}
		mu.Lock()
		defer mu.Unlock()

		for _, existing := range authors {
			if existing == in {
				return false
			}
		}
		return true
	}
	addAuthor := func(in string) {
		mu.Lock()
		defer mu.Unlock()
		authors = append(authors, in)
	}
	addPR := func(in prInfo) {
		mu.Lock()
		defer mu.Unlock()
		if in.Labels.needsDetails() {
			prNeedsDetails = append(prNeedsDetails, in)
		} else if in.Labels.needsToList() {
			prInfos = append(prInfos, in)
		}
	}

	for i := 0; i < numberOfThreads; i++ {
		wg.Add(1)
		go func() {
			// load meta data about PRs
			defer wg.Done()

			for b := range prChan {
				fmt.Print(".")

				if b.isPR {
					prInfo, err := loadPRInfo(b.key)
					if err != nil {
						errChan <- err
						break
					}
					addPR(prInfo)
					continue
				}
				author, err := loadAuthorInfo(b.key)
				if err != nil {
					errChan <- err
					break
				}
				if shouldLoad(author) {
					addAuthor(author)
				}

			}
		}()
	}

	go func() {
		// wait for the loading to finish
		wg.Wait()
		close(wgDone)
	}()

	var err error
	select {
	case <-wgDone:
		break
	case err = <-errChan:
		break
	}

	fmt.Println()

	sort.Strings(authors)

	return prInfos, prNeedsDetails, authors, err
}

func groupPRs(prInfos []prInfo) prsByType {
	prPerType := prsByType{}

	for _, info := range prInfos {
		var typ, component string
		for _, lbl := range info.Labels {
			switch {
			case strings.HasPrefix(lbl.Name, prefixType):
				typ = strings.TrimPrefix(lbl.Name, prefixType)
			case strings.HasPrefix(lbl.Name, prefixComponent):
				component = strings.TrimPrefix(lbl.Name, prefixComponent)
			}
		}
		switch typ {
		case "":
			typ = "Other"
		case "Bug":
			typ = "Bug fixes"
		}

		if component == "" {
			component = "Other"
		}
		components, exists := prPerType[typ]
		if !exists {
			components = prsByComponent{}
			prPerType[typ] = components
		}

		prsPerComponentAndType := components[component]
		components[component] = append(prsPerComponentAndType, info)
	}
	return prPerType
}

func createSortedPrTypeSlice(prPerType prsByType) []sortedPRType {
	var data []sortedPRType
	for typeKey, typeElem := range prPerType {
		newPrType := sortedPRType{
			Name: typeKey,
		}
		for componentKey, prInfos := range typeElem {
			newComponent := sortedPRComponent{
				Name:    componentKey,
				PrInfos: prInfos,
			}
			sort.Slice(newComponent.PrInfos, func(i, j int) bool {
				return newComponent.PrInfos[i].Number < newComponent.PrInfos[j].Number
			})
			newPrType.Components = append(newPrType.Components, newComponent)
		}
		sort.Slice(newPrType.Components, func(i, j int) bool {
			return newPrType.Components[i].Name < newPrType.Components[j].Name
		})
		data = append(data, newPrType)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})
	return data
}

func releaseSummary(summaryFile string) (string, error) {
	contentSummary, err := ioutil.ReadFile(summaryFile)
	if err != nil {
		return "", err
	}
	return string(contentSummary), nil
}

func getOutput(fileout string) (*os.File, error) {
	if fileout == "" {
		return os.Stdout, nil
	}

	return os.OpenFile(fileout, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func getStringForPullRequestInfos(prPerType prsByType) (string, error) {
	data := createSortedPrTypeSlice(prPerType)

	t := template.Must(template.New("markdownTemplatePR").Parse(markdownTemplatePR))
	buff := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buff, "markdownTemplatePR", data); err != nil {
		return "", err
	}
	return buff.String(), nil
}

func getStringForKnownIssues(issues []knownIssue) (string, error) {
	if len(issues) == 0 {
		return "", nil
	}
	t := template.Must(template.New("markdownTemplateKnownIssues").Parse(markdownTemplateKnownIssues))
	buff := bytes.Buffer{}
	if err := t.ExecuteTemplate(&buff, "markdownTemplateKnownIssues", issues); err != nil {
		return "", err
	}
	return buff.String(), nil
}

func groupAndStringifyPullRequest(pr []prInfo) (string, error) {
	if len(pr) == 0 {
		return "", nil
	}
	prPerType := groupPRs(pr)
	prStr, err := getStringForPullRequestInfos(prPerType)
	if err != nil {
		return "", err
	}
	return prStr, nil
}

func main() {
	from := flag.String("from", "", "from sha/tag/branch")
	to := flag.String("to", "HEAD", "to sha/tag/branch")
	versionName := flag.String("version", "", "name of the version (has to be the following format: v11.0.0)")
	summaryFile := flag.String("summary", "", "readme file on which there is a summary of the release")
	fileout := flag.String("file", "", "file on which to write release notes, stdout if empty")
	flag.Parse()

	releaseNotes := releaseNote{
		Version: *versionName,
	}

	// summary of the release
	if *summaryFile != "" {
		summary, err := releaseSummary(*summaryFile)
		if err != nil {
			log.Fatal(err)
		}
		releaseNotes.Announcement = summary
	}

	// known issues
	if *versionName != "" {
		knownIssues, err := loadKnownIssues(*versionName)
		if err != nil {
			log.Fatal(err)
		}
		knownIssuesStr, err := getStringForKnownIssues(knownIssues)
		if err != nil {
			log.Fatal(err)
		}
		releaseNotes.KnownIssues = knownIssuesStr
	}

	// changelog with pull requests
	prs, authorCommits, commits, err := loadMergedPRs(*from, *to)
	if err != nil {
		log.Fatal(err)
	}
	prInfos, prNeedsDetails, authors, err := loadAllPRs(prs, authorCommits)
	if err != nil {
		log.Fatal(err)
	}
	releaseNotes.ChangeLog, err = groupAndStringifyPullRequest(prInfos)
	if err != nil {
		log.Fatal(err)
	}
	releaseNotes.AddDetails, err = groupAndStringifyPullRequest(prNeedsDetails)
	if err != nil {
		log.Fatal(err)
	}

	// changelog metrics
	if commits > 0 && len(authors) > 0 {
		releaseNotes.ChangeMetrics = fmt.Sprintf(`
The release includes %d commits (excluding merges)

Thanks to all our contributors: @%s
`, commits, strings.Join(authors, ", @"))
	}

	out, err := getOutput(*fileout)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if err := releaseNotes.generate(out); err != nil {
		log.Fatal(err)
	}
}
