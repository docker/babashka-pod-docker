package docker

import (
	"github.com/docker/distribution/reference"
	"github.com/kballard/go-shellquote"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/patternmatcher"
	"github.com/moby/patternmatcher/ignorefile"

	//"reflect"
	"crypto/sha256"
	"crypto/sha512"

	"encoding/json"
	"fmt"
	"strings"

	"babashka-pod-docker/babashka"
)

type Reference struct {
	Path   string `json:"path"`
	Domain string `json:"domain,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Digest string `json:"digest,omitempty"`
}

type Error struct {
	Error string `json:"error"`
}

type Ignore struct {
	Patterns []string `json:"patterns"`
}

func patterns(s string) (Ignore, error) {
	patterns, err := ignorefile.ReadAll(strings.NewReader(s))
	if err != nil {
		return Ignore{}, err
	}
	return Ignore{Patterns: patterns}, err
}

func matches(path string, patterns []string) (bool, error) {
	return patternmatcher.MatchesOrParentMatches(path, patterns)
}

func parse_uri(s string) (Reference, error) {
	tag, domain, path, digest := "", "", "", ""

	sha256.New()
	sha512.New()

	ref, err := reference.Parse(s)
	if err != nil {
		return Reference{}, err
	}
	//fmt.Printf("%s\n", reflect.TypeOf(ref));

	if tagged, ok := ref.(reference.NamedTagged); ok {
		tag = tagged.Tag()
	}
	if named, ok := ref.(reference.Named); ok {
		domain = reference.Domain(named)
		path = reference.Path(named)
	}
	if digested, ok := ref.(reference.Canonical); ok {
		digest = digested.Digest().String()
	}
	//u, err := json.Marshal(Reference{Path: path, Domain: domain, Tag: tag, Digest: digest})
	return Reference{Path: path, Domain: domain, Tag: tag, Digest: digest}, err
}

func ProcessMessage(message *babashka.Message) (any, error) {
	switch message.Op {
	case "describe":
		return &babashka.DescribeResponse{
			Format: "json",
			Namespaces: []babashka.Namespace{
				{
					// this is the pod-id
					Name: "docker.tools",
					Vars: []babashka.Var{
						{
							Name: "parse-image-name",
						},
						{
							Name: "parse-dockerfile",
						},
						{
							Name: "parse-shellwords",
						},
						{
							Name: "dockerignore-patterns",
						},
						{
							Name: "dockerignore-matches",
						},
					},
				},
			},
		}, nil
	case "invoke":
		switch message.Var {
		case "docker.tools/parse-image-name":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			return parse_uri(args[0])
		case "docker.tools/parse-dockerfile":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}
			reader := strings.NewReader(args[0])
			return parser.Parse(reader)
		case "docker.tools/parse-shellwords":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}
			return shellquote.Split(args[0])

		case "docker.tools/dockerignore-patterns":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			return patterns(args[0])

		case "docker.tools/dockerignore-matches":
			type MyType struct {
				Path     string   `json:"path"`
				Patterns []string `json:"patterns"`
			}
			args := []MyType{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			return matches(args[0].Path, args[0].Patterns)

		default:
			return nil, fmt.Errorf("Unknown var %s", message.Var)
		}
	default:
		return nil, fmt.Errorf("Unknown op %s", message.Op)
	}
}
