package manifest

import (
	"log"
	"strings"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"encoding/base64"

	"gopkg.in/yaml.v2"

	"../github"
	"../utils"
)

type (
	Manifest struct {
		Info Info `yaml:"info"`
	}

	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	ManifestHandler struct{}
)

func (mh ManifestHandler) Handle(event github.Push) error {
	modified := false

	for _, commit := range event.Commits {
		log.Println("Changes: ", append(commit.Added, commit.Modified...))

		if utils.Contains("manifest.yml", append(commit.Added, commit.Modified...)) {
			modified = true
		}
	}

	if modified {
		resp, err := http.Get(strings.Replace(event.Repository.ContentUrl, "{+path}", "manifest.yml", 1))
		defer resp.Body.Close()

		if err != nil {
			return err
		}

		fileContent := &github.FileContent{}
		data, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(data, fileContent)
		if err != nil {
			return err
		}

		data, err = base64.StdEncoding.DecodeString(fileContent.Content)
		if err != nil {
			return err
		}

		manifest := &Manifest{}
		yaml.Unmarshal(data, manifest)

		log.Println(manifest)
	}

	return nil
}
