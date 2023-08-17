package docker

import (
	"github.com/docker/distribution/reference"
	"github.com/docker/index-cli-plugin/lsp"
	"github.com/docker/scout-cli-plugin/sbom"
	"github.com/kballard/go-shellquote"
	"github.com/moby/buildkit/frontend/dockerfile/parser"

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

func generate_sbom(message *babashka.Message, image string, username string, password string) error {
	tx_channel := make(chan string)

	go func() error {
		for {
			tx, ok := <-tx_channel
			if ok && tx != "" {
				err := babashka.WriteNotDoneInvokeResponse(message, tx)
				if err != nil {
					babashka.WriteErrorResponse(message, err)
				}
			} else {
				tx_channel = nil
				break
			}
		}
		babashka.WriteInvokeResponse(message, "done")
		return nil
	}()

	l := lsp.New()

	if username != "" && password != "" {
		l.WithAuth(username, password)
	}

	return l.Send(image, tx_channel)
}

func generate_hashes(message *babashka.Message, s string) error {
	tx_channel := make(chan string)

	go func() error {
		for {
			tx := <-tx_channel
			if tx != "" {
				err := babashka.WriteNotDoneInvokeResponse(message, tx)
				if err != nil {
					babashka.WriteErrorResponse(message, err)
				}

			} else {
				break
			}
		}
		return nil
	}()

	return lsp.New().SendFileHashes(s, tx_channel)
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
							Name: "sbom",
							Code: `
(defn sbom
  ([image cb]
   (sbom image cb {}))
  ([image cb opts]
   (babashka.pods/invoke
     "docker.tools"
     'docker.tools/generate-sbom
     [image]
     {:handlers {:success (fn [event]
                            (cb event))
                 :error   (fn [{:keys [:ex-message :ex-data]}]
                            (binding [*out* *err*]
                              (println "ERROR:" ex-message)))
		 :done    (fn [] (cb "done"))}})))`,
						},
						{
							Name: "hashes",
							Code: `
(defn hashes
  ([image cb]
   (hashes image cb {}))
  ([image cb opts]
   (babashka.pods/invoke
     "docker.tools"
     'docker.tools/generate-hashes
     [image]
     {:handlers {:success (fn [event]
                            (cb event))
                 :error   (fn [{:keys [:ex-message :ex-data]}]
                            (binding [*out* *err*]
                              (println "ERROR:" ex-message)))
			      :done    (fn [] (cb {:status "done"}))}})))`,
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
		case "docker.tools/generate-sbom":
			args := []string{}

			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}
			if len(args) == 3 {
				err := generate_sbom(message, args[0], args[1], args[2])
				if err != nil {
					babashka.WriteErrorResponse(message, err)
				}
			} else {
				err := generate_sbom(message, args[0], "", "")
				if err != nil {
					babashka.WriteErrorResponse(message, err)
				}
			}
			return "running", nil

		case "docker.tools/generate-hashes":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			err := generate_hashes(message, args[0])
			if err != nil {
				babashka.WriteErrorResponse(message, err)
			}

			return "done", nil
		case "docker.tools/scout-push":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			err := sbom.NewIndexer()
			if err != nil {
				babashka.WriteErrorResponse(message, err)
			}

			return "done", nil

		default:
			return nil, fmt.Errorf("Unknown var %s", message.Var)
		}
	default:
		return nil, fmt.Errorf("Unknown op %s", message.Op)
	}
}
