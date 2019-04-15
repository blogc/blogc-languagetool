package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	reStripNewlines = regexp.MustCompile(` *\r?\n *`)
)

func walkNode(node *html.Node, insideBlock bool) ([]string, error) {
	a := node.DataAtom

	if node.Type != html.ElementNode && node.Type != html.TextNode {
		return nil, fmt.Errorf("html: unexpected element block type: %s", node.Data)
	}

	if node.Type == html.TextNode {
		if insideBlock {
			return []string{node.Data}, nil
		}

		return []string{}, nil
	}

	if a == atom.Html {
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			if n.DataAtom != atom.Body {
				continue
			}

			tmp, err := walkNode(n, false)
			if err != nil {
				return nil, err
			}

			return tmp, nil
		}

		return nil, fmt.Errorf("html: failed to find body")
	}

	if a == atom.Body || a == atom.Blockquote || a == atom.Ul || a == atom.Ol {
		rv := []string{}

		for n := node.FirstChild; n != nil; n = n.NextSibling {
			tmp, err := walkNode(n, false)
			if err != nil {
				return nil, err
			}

			rv = append(rv, tmp...)
		}

		return rv, nil
	}

	isBlock := a == atom.P || a == atom.Li || a == atom.H1 || a == atom.H2 || a == atom.H3 || a == atom.H4 || a == atom.H5 || a == atom.H6
	isInline := a == atom.Em || a == atom.Strong || a == atom.Code || a == atom.A // || atom.Img

	if isBlock || isInline {
		t := []string{}

		for n := node.FirstChild; n != nil; n = n.NextSibling {
			tmp, err := walkNode(n, true)
			if err != nil {
				return nil, err
			}

			t = append(t, tmp...)
		}

		rv := strings.Join(t, "")

		if isBlock {
			return []string{strings.TrimSpace(reStripNewlines.ReplaceAllString(rv, " "))}, nil
		}

		return []string{rv}, nil
	}

	return []string{}, nil
}

func walk(doc *html.Node) ([]string, error) {
	if doc.Type != html.DocumentNode {
		return nil, fmt.Errorf("html: root element must be DocumentNode")
	}

	rv := []string{}

	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type != html.ElementNode {
			return nil, fmt.Errorf("html: unexpected block type in root element: %s", node.Data)
		}

		nodes, err := walkNode(node, false)
		if err != nil {
			return nil, err
		}

		rv = append(rv, nodes...)
	}

	return rv, nil
}

func html2text(content string) (string, error) {
	logrus.Info("Converting HTML to text")

	buffer := strings.NewReader(content)

	doc, err := html.Parse(buffer)
	if err != nil {
		return "", nil
	}

	blocks, err := walk(doc)
	if err != nil {
		return "", nil
	}

	return strings.Join(blocks, "\n\n"), nil
}
