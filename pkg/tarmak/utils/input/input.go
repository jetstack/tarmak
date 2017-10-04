package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/cli"
)

var RegexpName = regexp.MustCompile("^[a-z0-9-]+$")
var RegexpDNS = regexp.MustCompile("^[a-z0-9-\\.]+$")

type Input struct {
	output io.Writer
	input  io.Reader

	ui cli.Ui

	stopCh  chan struct{}
	inputCh chan string

	Prompt string
}

type AskOpen struct {
	Query      string
	AllowEmpty bool
	Default    string
}

func (q *AskOpen) Question() string {
	output := []string{q.Query}
	if q.Default != "" {
		output[0] += fmt.Sprintf(" (default '%s')", q.Default)
	}
	output = append(output, ">")
	return strings.Join(output, "\n")
}

type AskSelection struct {
	Query   string
	Choices []string
	Default int
}

func (q *AskSelection) Question() string {
	output := []string{q.Query}
	for pos, choice := range q.Choices {
		var defaultText string
		if pos == q.Default {
			defaultText = " (default)"
		}
		output = append(output, fmt.Sprintf("% 3d) %s%s", pos+1, choice, defaultText))
	}
	output = append(output, ">")
	return strings.Join(output, "\n")
}

type AskMultipleSelection struct {
	AskSelection    *AskSelection
	SelectedChoices []bool
	MinSelected     int
	MaxSelected     int
}

func (q *AskMultipleSelection) Question() string {
	var output []string

	for pos, choice := range q.AskSelection.Choices {
		var selected string
		if q.SelectedChoices[pos] {
			selected = " (selected)"
		}

		output = append(output, fmt.Sprintf("% 3d) %s%s", pos+1, choice, selected))
	}

	output = append(output, fmt.Sprintf("% 3d) continue with selection", len(q.AskSelection.Choices)+1))
	output = append(output, ">")

	return strings.Join(output, "\n")
}

type AskYesNo struct {
	Query   string
	Default bool
}

func (q *AskYesNo) Question() string {
	return fmt.Sprintf("%s %s", q.Query, q.Option())
}

func (q *AskYesNo) Option() string {
	if q.Default {
		return "[Y/n]"
	}
	return "[y/N]"
}

func New(i io.Reader, o io.Writer) *Input {
	input := &Input{
		input:  i,
		output: o,
		stopCh: make(chan struct{}),
	}
	input.initUI()
	return input
}

func (i *Input) initUI() {
	i.ui = &cli.ConcurrentUi{
		Ui: &cli.ColoredUi{
			ErrorColor:  cli.UiColorRed,
			WarnColor:   cli.UiColorYellow,
			InfoColor:   cli.UiColorBlue,
			OutputColor: cli.UiColorNone,
			Ui: &cli.BasicUi{
				ErrorWriter: i.output,
				Writer:      i.output,
				Reader:      i.input,
			},
		},
	}
}

func (i *Input) Close() {
	close(i.stopCh)
}

func (i *Input) startListening() {
	i.stopCh = make(chan struct{})
	i.inputCh = make(chan string)

	go func(ch chan string) {
		reader := bufio.NewReader(i.input)
		for {
			s, err := reader.ReadString('\n')
			if err != nil { // Maybe log non io.EOF errors, if you want
				close(ch)
				return
			}
			ch <- s
		}
	}(i.inputCh)
}

func (i *Input) stopListening() {
	close(i.stopCh)
}

func (i *Input) Warn(a ...interface{}) {
	i.ui.Warn(fmt.Sprint(a...))
}

func (i *Input) Warnf(format string, a ...interface{}) {
	i.ui.Warn(fmt.Sprintf(format, a...))
}

func (i *Input) Askf(format string, a ...interface{}) (string, error) {
	return i.ui.Ask(fmt.Sprintf(format, a...))
}

func (i *Input) AskYesNo(question *AskYesNo) (bool, error) {

	for {
		response, err := i.Askf(question.Question())
		if err != nil {
			return false, err
		}

		res := strings.ToLower(response)
		if res == "y" || res == "yes" {
			return true, nil
		} else if res == "n" || res == "no" {
			return false, nil
		} else if res == "" {
			break
		} else {
			i.Warnf("bad response %s", question.Option())
			continue
		}
	}

	return question.Default, nil
}

func (i *Input) AskSelection(question *AskSelection) (int, error) {

	for {
		response, err := i.Askf(question.Question())
		if err != nil {
			return -1, err
		}

		if response == "" {
			if question.Default >= 0 {
				break
			} else {
				i.Warn("nothing entered and no default set")
			}
		} else if n, err := strconv.Atoi(response); err != nil || n < 0 || n > len(question.Choices) {
			i.Warnf("response must be a number between 1 and %d\n", len(question.Choices))
		} else {
			return n - 1, nil
		}

	}
	return question.Default, nil
}

func (i *Input) AskMultipleSelection(question *AskMultipleSelection) (responseSlice []string, err error) {

	if len(question.SelectedChoices) != len(question.AskSelection.Choices) {
		return []string{}, errors.New("length of choice and selected slice does not match")
	}

	i.ui.Output(question.AskSelection.Query)

	choiceSize := len(question.SelectedChoices)

	for {

		response, err := i.Askf(question.Question())
		if err != nil {
			return []string{}, err
		}

		if response != "" {
			if cRange, err := question.parseResponce(response); err != nil || cRange[0] < 1 || cRange[1] == choiceSize+1 && cRange[0] != cRange[1] || cRange[1] > choiceSize+1 {
				i.Warnf("response must be a range between 1 and %d or a number between 1 and %d\n", choiceSize, choiceSize+1)

			} else if cRange[0] == cRange[1] && cRange[0] == choiceSize+1 {
				if question.insideMinMax() {
					responseSlice = question.returnSelected()
					break

				} else {
					i.Warn(fmt.Sprintf("please select between %d and %d choices", question.MinSelected, question.MaxSelected))
				}

			} else {
				if cRange[0] == cRange[1] {
					question.SelectedChoices[cRange[0]-1] = !question.SelectedChoices[cRange[0]-1]
				} else {
					for n := cRange[0] - 1; n < cRange[1]; n++ {
						question.SelectedChoices[n] = true
					}
				}
			}

		} else {
			if question.insideMinMax() {
				responseSlice = question.returnSelected()
				break

			} else {
				i.Warn(fmt.Sprintf("please select between %d and %d choices", question.MinSelected, question.MaxSelected))
			}
		}
	}

	return responseSlice, nil
}

func (q *AskMultipleSelection) returnSelected() (responseSlice []string) {
	for pos, choice := range q.AskSelection.Choices {
		if q.SelectedChoices[pos] {
			responseSlice = append(responseSlice, choice)
		}
	}

	return responseSlice
}

func (q *AskMultipleSelection) insideMinMax() bool {
	var count int
	for _, choice := range q.SelectedChoices {
		if choice {
			count++
		}
	}

	if count < q.MinSelected || count > q.MaxSelected {
		return false
	}

	return true
}

// Get selections
func (q *AskMultipleSelection) parseResponce(response string) (selects []int, err error) {
	choiceRange := []int{1, len(q.AskSelection.Choices)}
	regex := regexp.MustCompile("^(([0-9]+(-[0-9]*)?)|(([0-9]*-)?[0-9]+))$")

	match := string(regex.Find([]byte(response)))
	if match == "" {
		return selects, fmt.Errorf("could not match: %s", response)
	}

	split := strings.Split(response, "-")

	if len(split) == 1 {
		n, err := strconv.Atoi(response)
		if err != nil {
			return selects, nil
		}
		choiceRange[0] = n
		choiceRange[1] = n

	} else {
		if split[0] != "" {
			n, err := strconv.Atoi(split[0])
			if err != nil {
				return selects, nil
			}
			choiceRange[0] = n
		}

		if split[1] != "" {
			n, err := strconv.Atoi(split[1])
			if err != nil {
				return selects, nil
			}
			choiceRange[1] = n
		}
	}

	if choiceRange[0] > choiceRange[1] {
		temp := choiceRange[1]
		choiceRange[1] = choiceRange[0]
		choiceRange[0] = temp
	}

	return choiceRange, nil
}

func (i *Input) AskOpen(question *AskOpen) (response string, err error) {

	for {
		response, err := i.Askf(question.Question())
		if err != nil {
			return "", err
		}

		if response == "" {
			if question.AllowEmpty || question.Default != "" {
				break
			} else {
				i.Warn("nothing entered, empty response not allowed")
			}
		} else {
			return response, nil
		}
	}
	return question.Default, nil

}
