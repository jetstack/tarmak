package utils

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func ListParameters(out io.Writer, keys []string, varMaps []map[string]string) {

	inlistMap := map[string]bool{}
	for _, key := range keys {
		inlistMap[key] = true
	}

	// populate map and list with all possible keys
	for _, varMap := range varMaps {
		for key, _ := range varMap {
			if _, ok := inlistMap[key]; !ok {
				keys = append(keys, key)
				inlistMap[key] = true
			}
		}
	}

	keysHeader := make([]interface{}, len(keys))
	for pos, _ := range keys {
		keysHeader[pos] = strings.ToUpper(keys[pos])
	}

	// init tab writter
	w := new(tabwriter.Writer)
	w.Init(out, 0, 8, 0, '\t', 0)

	// output headers for tab list
	formatString := strings.Repeat("%s\t", len(keys)) + "\n"
	fmt.Fprintf(
		w,
		formatString,
		keysHeader...,
	)

	for _, varMap := range varMaps {
		fields := make([]interface{}, len(keys))
		for pos, key := range keys {
			if val, ok := varMap[key]; ok {
				fields[pos] = val
			} else {
				fields[pos] = ""
			}
		}
		fmt.Fprintf(
			w,
			formatString,
			fields...,
		)
	}

	w.Flush()

}
