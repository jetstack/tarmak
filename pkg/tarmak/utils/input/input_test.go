package input_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

func Test_Input_AskYesNo_Yes(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskYesNo{
		Query:   "Should this test pass?",
		Default: true,
	}

	go func() {
		inWriter.Write([]byte("Y\n"))
		//inWriter.Close()
	}()

	resp, err := i.AskYesNo(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if !resp {
		t.Error("expected true response")
	}
}

func Test_Input_AskYesNo_No(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskYesNo{
		Query:   "Should this test pass?",
		Default: true,
	}

	go func() {
		inWriter.Write([]byte("no\n"))
		//inWriter.Close()
	}()

	resp, err := i.AskYesNo(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if resp {
		t.Error("expected false response")
	}
}

func Test_Input_AskYesNo_Default_No(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskYesNo{
		Query:   "Should this test pass?",
		Default: false,
	}

	go func() {
		inWriter.Write([]byte("\n"))
	}()

	resp, err := i.AskYesNo(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if resp {
		t.Error("expected false response")
	}
}

func Test_Input_AskYesNo_Wrong_Reask_Yes(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskYesNo{
		Query:   "Should this test pass?",
		Default: false,
	}

	go func() {
		inWriter.Write([]byte("xxx\n"))
		inWriter.Write([]byte("Y\n"))
	}()

	resp, err := i.AskYesNo(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if !resp {
		t.Error("expected true response")
	}
}

func Test_Input_AskOpen_NoDefault_NoResponse(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskOpen{
		Query: "Should this test pass?",
	}

	go func() {
		inWriter.Write([]byte("\n"))
		inWriter.Write([]byte("no_default\n"))
	}()

	resp, err := i.AskOpen(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "no_default", resp; exp != act {
		t.Errorf("unexpected response, exp=%s act=%s", exp, act)
	}
}

func Test_Input_AskOpen_Default_NoResponse(t *testing.T) {
	out := new(bytes.Buffer)
	inReader, inWriter := io.Pipe()
	i := input.New(inReader, out)

	question := &input.AskOpen{
		Query:   "Should this test pass?",
		Default: "no never",
	}

	go func() {
		inWriter.Write([]byte("\n"))
	}()

	resp, err := i.AskOpen(question)
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if exp, act := "no never", resp; exp != act {
		t.Errorf("unexpected response, exp=%s act=%s", exp, act)
	}
}
