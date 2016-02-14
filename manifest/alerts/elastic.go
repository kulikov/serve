package alerts

import (
	"log"
	"gopkg.in/yaml.v2"
	"../../manifest"
	"fmt"
	"strings"
	"regexp"
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

func (ea ElasticAlertPlugin) Run(mft *manifest.Manifest) error {
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
                "elastic",
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


//class ElasticAlert(BasicAlert):
//    file_name = "check_mk_elastic_alerts.sh"
//
//    _elastic_nodes = {
//        "qa": "elastic.qa.inn.ru",
//        "live": "elastic.lux.inn.eu"
//    }
//
//    def register(self, alert, service):
//        warn = self._threshold(alert['elastic'], "warn")
//        crit = self._threshold(alert['elastic'], "crit")
//
//        if warn or crit:
//            self.checks += (
//                "result=$(check_json.pl %s %s "
//                "--url 'http://%s:9200/logstash-*/_search?q="
//                "(%s) AND timemillis:['$(( ($(date +%%s) * 1000) - %d ))' TO '$(($(date +%%s) * 1000))']&search_type=count' "
//                "--attribute '{hits}->{total}' "
//                "--perfvars '{hits}->{total}') \n"
//                """echo "$? services.%s.%s perfdata=$(echo $result | sed 's/.*- total: \([0-9]*\).*/\\1/').0 """
//                """total=$(echo $result | sed 's/.*total=\(.*\).*/\\1/') by query '%s';" \n\n"""
//            ) % (
//                warn,
//                crit,
//                self._env_val(self._elastic_nodes),
//                alert['elastic']['query'].replace("'", "'\\''"),
//                self._parse_duration_ms(alert['elastic'].get('from', '15min')),
//                service['name'],
//                re.sub('\W+', '-', alert.get('name', 'elastic')).lower(),
//                re.sub("[^\w\s:\-\.\(\)]+", '', alert['elastic']['query'])
//            )
//
//    def _threshold(self, obj, level):
//        _min = self._env_val(obj.get(level + ".min", ""))
//        _max = self._env_val(obj.get(level + ".max", obj.get(level, "")))
//
//        if _min or _max:
//            return "-%s %s:%s" % (level[0].lower(), _min, _max)
