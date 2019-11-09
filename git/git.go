package git

import (
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

var ErrOriginDNE = errors.New("origin url does not exist")

var originRe = regexp.MustCompile(`(?:/|:)(\w+/\w+)\.git`)

func GetOrigin(gitFile string) (string, error) {
	return getOrigin(gitFile)
}

func getOrigin(gitFile interface{}) (string, error) {
	g, err := ini.Load(gitFile)
	if err != nil {
		return "", errors.Wrap(err, "Failed to read .git/config")
	}

	url := g.Section(`remote "origin"`).Key("url").String()
	if url == "" {
		return "", ErrOriginDNE
	}

	matches := originRe.FindAllStringSubmatch(url, 1)
	if len(matches) != 1 || len(matches[0]) != 2 {
		return "", ErrOriginDNE
	}

	return matches[0][1], nil
}
