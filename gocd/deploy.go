package gocd

import (
	"encoding/json"

	"github.com/spf13/viper"

	"../manifest"
	"log"
)

type DeployPlugin struct{}

func (ea DeployPlugin) Run(conf *viper.Viper, mft *manifest.Manifest) error {
	pipeline := Pipeline{
		Name: mft.Info.Name,
		Materials: []Material{
			Material{
				Type: "git",
				Attributes: MaterialAttrs{
					Url: mft.GitSshUrl,
					Destination: "sources",
				},
			},
		},
		EnvVariables: []EnvVariable{
			EnvVariable{
				Name: "SHA",
				Value: "",
				Secure: false,
			},
			EnvVariable{
				Name: "BRANCH",
				Value: "master",
				Secure: false,
			},
			EnvVariable{
				Name: "GITHUB_TOKEN",
				EncryptedValue: "XXX", // @todo: move to config by github domain
				Secure: true,
			},
		},
	}

	bytes, err := json.Marshal(pipeline)

	if err != nil {
		return err
	}

	log.Println(string(bytes))

	return nil
}
