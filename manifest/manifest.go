package manifest

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/kulikov/serve/github"
	"github.com/kulikov/serve/utils"
)

type (
	Manifest struct {
		Sha       string
		GitSshUrl string
		Source    []byte

		Info         Info         `yaml:"info"`
		Notification Notification `yaml:"notification"`
	}

	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Owner   Owner  `yaml:"owner"`
	}

	Owner struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	}

	Notification struct {
		Channel string `yaml:"channel"`
	}

	Plugin interface {
		Run(conf *viper.Viper, manf *Manifest) error
	}
)

func HandleGithubChanges(conf *viper.Viper, plugins []Plugin, payload string) error {
	event := &github.Push{}
	json.Unmarshal([]byte(payload), event)

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

		manf := &Manifest{Sha: file.Sha, GitSshUrl: event.Repository.SshUrl, Source: data}
		yaml.Unmarshal(data, manf)

		RunPlugins(conf, plugins, manf)
	} else {
		log.Println("manifest.yml not changed")
	}

	return nil
}

func InitConfig(configFile string) (*viper.Viper, error) {
	conf := viper.New()
	conf.SetConfigType("yml")

	for _, file := range strings.Split(configFile, ",") {
		conf.SetConfigFile(file)

		if err := conf.MergeInConfig(); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func RunPlugins(conf *viper.Viper, plugins []Plugin, manf *Manifest) {
	wg := sync.WaitGroup{}

	for _, plugin := range plugins {
		wg.Add(1)

		go func(p Plugin) {
			defer wg.Done()

			err := p.Run(conf, manf)

			if err != nil {
				log.Printf("%T: %s\n", p, err)
			}
		}(plugin)
	}

	wg.Wait()
}
