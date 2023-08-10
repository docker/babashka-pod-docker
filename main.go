package main

import (
	"babashka-pod-docker/babashka"
	"babashka-pod-docker/docker"
	"fmt"
	"os"

	"github.com/atomist-skills/go-skill"
	"github.com/sirupsen/logrus"
)

func main() {
	skill.Log.SetLevel(logrus.ErrorLevel)

	args := os.Args

	if len(args) < 2 {
		args = append(os.Args, "pod")
	}

	switch args[1] {

	case "docker-cli-plugin-metadata":
		metadata := `{"SchemaVersion": "0.1.0", "Vendor": "Docker Inc.", "Version": "v0.0.1", "ShortDescription": "Docker Pod"}`
		fmt.Println(metadata)

	case "pod":
		for {
			message, err := babashka.ReadMessage()
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				continue
			}

			res, err := docker.ProcessMessage(message)
			if err != nil {
				babashka.WriteErrorResponse(message, err)
				continue
			}

			describeres, ok := res.(*babashka.DescribeResponse)
			if ok {
				babashka.WriteDescribeResponse(describeres)
				continue
			}

			if res != "running" {
				babashka.WriteInvokeResponse(message, res)
			}
		}
	}
}
