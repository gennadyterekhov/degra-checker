package checker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"
)

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type violation struct {
	Edge
	FromLayer string `json:"fromLayer"`
	ToLayer   string `json:"toLayer"`
}

func ReadGraph(input io.Reader) ([]Edge, error) {
	decoder := json.NewDecoder(input)
	var data json.RawMessage
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return nil, errors.New("expected exactly one graph JSON document")
	}

	var edges []Edge
	switch data[0] {
	case '{':
		var grouped map[string][]string
		if err := json.Unmarshal(data, &grouped); err != nil {
			return nil, err
		}
		for from, targets := range grouped {
			if !validFile(from) || targets == nil {
				return nil, fmt.Errorf("expected a relative file path and an array of imports for %q", from)
			}
			for _, to := range targets {
				edges = append(edges, Edge{from, to})
			}
		}
	case '[':
		if err := json.Unmarshal(data, &edges); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("expected an import map or an array of edges")
	}

	for _, e := range edges {
		if !validFile(e.From) || !validFile(e.To) {
			return nil, fmt.Errorf("invalid file path in edge from %q to %q", e.From, e.To)
		}
	}
	slices.SortFunc(edges, func(a, b Edge) int {
		if order := strings.Compare(a.From, b.From); order != 0 {
			return order
		}
		return strings.Compare(a.To, b.To)
	})
	return slices.Compact(edges), nil
}

func validFile(name string) bool {
	return name != "" && name != "." && name == path.Clean(name) &&
		!path.IsAbs(name) && within(name, ".") && !strings.ContainsRune(name, 0)
}

func within(name, dir string) bool {
	if dir == "" || dir == "." {
		return name != ".." && !strings.HasPrefix(name, "../")
	}
	return name == dir || strings.HasPrefix(name, dir+"/")
}
