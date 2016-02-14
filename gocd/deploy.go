package gocd

import (
	"encoding/json"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"../manifest"
)

type (
	DeployPlugin struct{}

	DeployManifest struct {
		Build []DeployBuild `yaml:"build"`
	}

	DeployBuild struct {
		Debian *DeployDebian `yaml:"debian"`
	}

	DeployDebian struct{}
)

func (ea DeployPlugin) Run(conf *viper.Viper, mft *manifest.Manifest) error {
	depl := DeployManifest{}
	yaml.Unmarshal(mft.Source, &depl)

	pipeline := Pipeline{
		Name: mft.Info.Name,
		Materials: []Material{
			Material{
				Type: "git",
				Attributes: MaterialAttrs{
					Url:         mft.GitSshUrl,
					Destination: "sources",
				},
			},
		},
	}

	for _, build := range depl.Build {
		if build.Debian != nil {
			pipeline.Stages = append(pipeline.Stages, Stage{
				Name: "Build",
				CleanWorkingDirectory: true,
				FetchMaterials:        true,
				Jobs: []Job{
					Job{
						Name:      "Create-Package",
						Resources: []string{"Builder", "Debian"},
						Tasks: []Task{
							Task{
								Type: "exec",
								Attributes: TaskAttributes{
									RunIf:            []string{"passed"},
									WorkingDirectory: "sources",
									Command:          "/bin/bash",
									Arguments: []string{
										"-c",
										"/var/go/inn-ci-tools/build-package.sh --suffix=-master --patch-version=$GO_PIPELINE_LABEL --repo=" + mft.GitSshUrl + " --distribution=unstable",
									},
								},
							},
						},
					},
				},
			})
		}
	}

	bytes, err := json.Marshal(pipeline)

	if err != nil {
		return err
	}

	log.Println(string(bytes))

	return nil
}
