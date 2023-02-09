package main

import (
	"dockerfileparse/user/parser/babashka"
	"dockerfileparse/user/parser/docker"

	"github.com/atomist-skills/go-skill"
	"github.com/sirupsen/logrus"
)

func main() {
	skill.Log.SetLevel(logrus.ErrorLevel)
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
		// TODO don't write done responses when callback is running
		babashka.WriteInvokeResponse(message, res)
	}
}
