package main

import (
	"github.com/blogc/go-blogc"
	"github.com/sirupsen/logrus"
)

func blogcParse(fileName string) (string, error) {
	logrus.WithField("blogc_version", blogc.Version).Info("Converting source to HTML using blogc")

	entry := &blogc.BuildContext{
		InputFiles: []blogc.File{blogc.FilePath(fileName)},
	}

	text, _, err := entry.GetEvaluatedVariable("CONTENT")
	return text, err
}
