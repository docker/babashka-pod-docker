package docker

import (
	"os"

	"github.com/docker/distribution/reference"
	"github.com/docker/scout-cli-plugin/push"
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

func scout_push(message *babashka.Message, image string, organization string, authToken string, username string, password string) error {
	os.Setenv("DOCKER_SCOUT_REGISTRY_USER", username)
	os.Setenv("DOCKER_SCOUT_REGISTRY_PASSWORD", password)

	return push.Push(image, namespace, authToken)
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
							Name: "scout-push",
							Code: `
(defn scout-push
  ([image cb]
   (hashes image cb {}))
  ([image cb opts]
   (babashka.pods/invoke
     "docker.tools"
     'docker.tools/scout-push
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
		case "docker.tools/scout-push":
			args := []string{}
			if err := json.Unmarshal([]byte(message.Args), &args); err != nil {
				return nil, err
			}

			err := scout_push(message, args[0], args[1], args[2], args[3], args[4])
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
