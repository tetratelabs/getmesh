package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/tetratelabs/getmesh/cmd"
)

const basePrefix = "doc/en"

func main() {
	runCommandGen()
	runContributingGen()
}

var (
	commandPrefix          = filepath.Join(basePrefix, "getmesh-cli", "reference")
	contributingPrefix     = filepath.Join(basePrefix, "community")
	contributingAttributes = []struct {
		path, title, beginning string
	}{
		{path: filepath.Join(contributingPrefix, "/building-and-testing"), title: "Building and Testing", beginning: "## Building & Testing\n"},
		{path: filepath.Join(contributingPrefix, "/contributing"), title: "Contributing to getmesh", beginning: "## Contributing\n"},
		{path: filepath.Join(contributingPrefix, "/release"), title: "Release process", beginning: "## Release\n"},
	}
)

func runCommandGen() {
	root := cmd.NewRoot("", "")
	if err := os.MkdirAll(commandPrefix, 0755); err != nil {
		panic(err)
	}
	cmdWriteFile(root)
	for _, c := range root.Commands() {
		cmdWriteFile(c)
	}
}

func cmdWriteFile(c *cobra.Command) {
	// process root
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdownCustom(c, buf, cmdLinkHandler); err != nil {
		panic(err)
	}

	var prefix string
	if c.Name() != "getmesh" {
		prefix = filepath.Join(commandPrefix, "getmesh_"+c.Name())
	} else {
		// root cmd
		prefix = filepath.Join(commandPrefix, c.Name())
	}
	_ = os.MkdirAll(prefix, 0755)

	p := filepath.Join(prefix, "_index.md")
	f, err := os.Create(p)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// should be able to handle multiple depth subcommands if c is not root:
	// if c.Name() != "getmesh" { for _, r := range c.Commands() {...} }

	_, err = f.WriteString(cmdFormatDoc(c.Name(), buf.String()))
	if err != nil {
		panic(err)
	}
}

const headerTemplate = `---
title: "%s"
url: %s
---
`

func cmdFormatDoc(name, base string) string {
	// trim the prefix until we reach "Synopsis" section which corresponds to "Command.Long" field.
	reg := regexp.MustCompile(`(### Synopsis\n)`)
	pos := reg.FindStringSubmatchIndex(base)
	if len(pos) >= 2 {
		base = base[pos[1]:]
	}

	// replace "###" to "####" for better appearance
	reg = regexp.MustCompile("### ")
	base = reg.ReplaceAllString(base, "#### ")

	// append hugo header
	var title, url string
	if name != "getmesh" {
		title = "getmesh " + name
		url = cmdGetURL("getmesh_" + name)
	} else {
		// root cmd
		title = name
		url = cmdGetURL(name)
	}

	header := fmt.Sprintf(headerTemplate, title, url)
	return header + base
}

func cmdLinkHandler(name string) string {
	return cmdGetURL(strings.TrimSuffix(name, path.Ext(name)))
}

func cmdGetURL(in string) string {
	return "/getmesh-cli/reference/" + strings.ToLower(in) + "/"
}

func runContributingGen() {
	raw, err := ioutil.ReadFile("CONTRIBUTING.md")
	if err != nil {
		panic(err)
	}

	original := string(raw)
	split := make([]int, len(contributingAttributes)+1)
	for i := 0; i < len(contributingAttributes)+1; i++ {
		if len(contributingAttributes) == i {
			split[i] = len(original)
		} else {
			split[i] = strings.Index(original, contributingAttributes[i].beginning)
		}
	}

	for i, attr := range contributingAttributes {
		if err := os.MkdirAll(attr.path, 0755); err != nil {
			panic(err)
		}

		p := filepath.Join(attr.path, "_index.md")
		f, err := os.Create(p)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		header := fmt.Sprintf(headerTemplate, attr.title, strings.TrimPrefix(attr.path, basePrefix))

		body := original[split[i]:split[i+1]]
		body = body[len(attr.beginning):] // trim the title since it overlaps with the title in header
		_, err = f.WriteString(header + body)
		if err != nil {
			panic(err)
		}
	}
}
