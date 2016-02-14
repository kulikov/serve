package alerts

import (
	"log"
	"fmt"
	"strings"
	"regexp"

	"gopkg.in/yaml.v2"
	"github.com/spf13/viper"

	"../../manifest"
)

type (
	ElasticAlertPlugin struct{}

	ElasticManifest struct {
		Alerts []ElasticAlerts `yaml:"alerts"`
	}

	ElasticAlerts struct {
		Name    string `yaml:"name"`
		Elastic *ElasticAlert `yaml:"elastic"`
	}

	ElasticAlert struct {
		Query string `yaml:"query"`
		Warn  interface{} `yaml:"warn"`
		Crit  interface{} `yaml:"crit"`
	}
)

func (ea ElasticAlertPlugin) Run(conf *viper.Viper, mft *manifest.Manifest) error {
	elmft := ElasticManifest{}
	yaml.Unmarshal(mft.Source, &elmft)

	checks := make([]string, len(elmft.Alerts))

	for _, alert := range elmft.Alerts {
		if el := alert.Elastic; el != nil {
			log.Println(el)

			checks = append(checks, fmt.Sprintf(
				"result=$(check_json.pl %v %v " +
				"--url 'http://%s:9200/logstash-*/_search?q=" +
				"(%s) AND timemillis:['$(( ($(date +%%s) * 1000) - %v ))' TO '$(($(date +%%s) * 1000))']&search_type=count' " +
				"--attribute '{hits}->{total}' " +
				"--perfvars '{hits}->{total}') \n" +
				`echo "$? services.%s.%s perfdata=$(echo $result | sed 's/.*- total: \([0-9]*\).*/\\1/').0 ` +
				`total=$(echo $result | sed 's/.*total=\(.*\).*/\\1/') by query '%s';" \n\n`,
				el.Warn,
				el.Crit,
				"elastic." + conf.GetString("env"),
				strings.Replace(el.Query, "'", "'\\''", -1),
				900000,
				mft.Info.Name,
				regexp.MustCompile(`\W+`).ReplaceAllString(strings.ToLower(alert.Name), "-"),
				regexp.MustCompile(`[^\w\s:\-\.\(\)]+`).ReplaceAllString(el.Query, ""),
			))
		}
	}

	log.Println(checks)

	return nil
}
