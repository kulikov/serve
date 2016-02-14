package manifest

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"../github"
	"../utils"
)

type (
	Manifest struct {
		Sha       string
		GitSshUrl string
		Source    []byte
		Info      Info `yaml:"info"`
	}

	Info struct {
		Name    string   `yaml:"name"`
		Version string   `yaml:"version"`
		Tags    []string `yaml:"tags"`
		Owner   Owner    `yaml:"owner"`
	}

	Owner struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	}

	Plugin interface {
		Run(conf *viper.Viper, mft *Manifest) error
	}

	ManifestHandler struct {
		Plugins []Plugin
	}
)

func (mh ManifestHandler) Handle(conf *viper.Viper, event github.Push) error {
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

		file := &github.FileContent{}
		data, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(data, file)
		if err != nil {
			return err
		}

		data, err = base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return err
		}

		mft := &Manifest{Sha: file.Sha, GitSshUrl: event.Repository.SshUrl, Source: data}
		yaml.Unmarshal(data, mft)

		mh.RunPlugins(conf, mft)
	}

	return nil
}

func (mh ManifestHandler) RunPlugins(conf *viper.Viper, mft *Manifest) {
	for _, plugin := range mh.Plugins {
		go func(p Plugin) {
			err := p.Run(conf, mft)

			if err != nil {
				log.Printf("%T: %s\n", p, err)
			}
		}(plugin)
	}
}
