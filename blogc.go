package main

import (
	"fmt"

	"github.com/blogc/go-blogc"
	"github.com/sirupsen/logrus"
)

func blogcParse(fileName string) (string, error) {
	logrus.WithField("blogc_version", blogc.Version).Info("Converting source to HTML using blogc")

	entry := &blogc.BuildContext{
		InputFiles: []blogc.File{blogc.FilePath(fileName)},
	}

	text, _, err := entry.GetEvaluatedVariable("CONTENT")
	if err != nil {
		return "", err
	}

	title, found, err := entry.GetEvaluatedVariable("TITLE")
	if err != nil {
		return "", err
	}
	if found {
		return fmt.Sprintf("<h1>%s</h1>\n%s", title, text), nil
	}

	return text, nil
}
