package gocd

import (
	"encoding/json"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"../manifest"
	"net/http"
)

type (
	DeployPlugin struct{}

	DeployManifest struct {
		Build  []DeployBuild `yaml:"build"`
		Deploy Deploy        `yaml:"deploy"`
	}

	DeployBuild struct {
		Debian *BuildDebian `yaml:"debian"`
	}

	BuildDebian struct{}

	Deploy struct {
		Debian *DeployDebian `yaml:"debian"`
	}

	DeployDebian struct {
		Cluster map[string]string `yaml:"cluster"`
	}
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

	if depl.Deploy.Debian != nil {
		deb := depl.Deploy.Debian
		project := "inn-" + mft.Info.Name
		qaNodes := deb.Cluster["qa"]
		liveNodes := deb.Cluster["live"]

		if qaNodes != "" {
			pipeline.Stages = append(pipeline.Stages, Stage{
				Name: "Deployment-QA",
				Jobs: []Job{
					Job{
						Name:      "Deploy",
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
										"dig +short " + qaNodes + " | sort | uniq | parallel -j 10 " +
											"/var/go/inn-ci-tools/go/go-package-deploy.sh --target={} --project=" + project +
											" --version=" + mft.Info.Version + ".$GO_PIPELINE_LABEL --purge-pattern=" + project + "-v.*",
									},
								},
							},
						},
					},
				},
			},
			Stage{
				Name: "Ready-to-Ship",
				Approval: Approval{
					Type: "manual",
				},
				Jobs: []Job{
					Job{
						Name:      "Approve-Package",
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
										"/var/go/inn-ci-tools/go/go-package-approve.sh --suffix=-master --patch-version=$GO_PIPELINE_LABEL " +
											"--project=" + project + " --version=" + mft.Info.Version + " --src-repo=unstable --dst-repo=stable",
									},
								},
							},
						},
					},
				},
			})
		}

		if liveNodes != "" {
			pipeline.Stages = append(pipeline.Stages, Stage{
				Name: "Deployment-Live",
				Approval: Approval{
					Type: "manual",
					Authorization: Authorization{
						Roles: []string{"QA"},
					},
				},
				Jobs: []Job{
					Job{
						Name:      "Deploy",
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
										"dig +short " + liveNodes + " | sort | uniq | parallel -j 10 " +
											"/var/go/inn-ci-tools/go/go-package-deploy.sh --target={} --project=" + project +
											" --version=" + mft.Info.Version + ".$GO_PIPELINE_LABEL --purge-pattern=" + project + "-v.*",
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

	req, _ := http.NewRequest("POST", "https://go.inn.ru/go/api/admin/pipelines", bytes)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.go.cd.v1+json")
	req.SetBasicAuth("login", "pass")

	resp, err := http.DefaultClient.Do(req)

	if (resp.StatusCode > 201) {
		return resp.Status
	}

	log.Println(string(bytes))

	return nil
}
