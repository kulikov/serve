package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/consul/api"
)

func main() {
	consul, _ := api.NewClient(api.DefaultConfig())

	splitRegex := regexp.MustCompile(":")
	upstreamRegex := regexp.MustCompile("[^\\w]+")

	upstreams := make(map[string][]map[string]interface{})
	servers := make(map[string]map[string]map[string]string)

	allServices, _, err := consul.Catalog().Services(&api.QueryOptions{})
	if err != nil {
		panic(err)
	}

	for s, tags := range allServices {
		hasHost := false
		for _, tag := range tags {
			if strings.HasPrefix(tag, "host:") {
				hasHost = true
				break
			}
		}

		if hasHost {
			services, _, err := consul.Health().Service(s, "", true, &api.QueryOptions{})
			if err != nil {
				panic(err)
			}

			for _, serv := range services {
				params := make(map[string]string)
				for _, t := range serv.Service.Tags {
					tt := splitRegex.Split(t, 2)
					if len(tt) > 1 {
						params[tt[0]] = tt[1]
					}
				}

				address := serv.Node.Address
				if serv.Service.Address != "" {
					address = serv.Service.Address
				}

				location, ok := params["location"]
				if ok == false {
					location = "/"
				}

				staging, ok := params["staging"]
				if ok == false {
					staging = "live"
				}

				upstream := upstreamRegex.ReplaceAllString("ups_"+params["host"]+"_"+location+"_"+staging, "_")

				if _, ok := upstreams[upstream]; ok == false {
					upstreams[upstream] = make([]map[string]interface{}, 0)
				}

				upstreams[upstream] = append(upstreams[upstream], map[string]interface{}{
					"address": address,
					"port":    serv.Service.Port,
				})

				if _, ok := servers[params["host"]]; ok == false {
					servers[params["host"]] = make(map[string]map[string]string, 0)
				}

				if _, ok := servers[params["host"]][location]; ok == false {
					servers[params["host"]][location] = make(map[string]string, 0)
				}

				if _, ok := servers[params["host"]][location][staging]; ok == false {
					servers[params["host"]][location][staging] = upstream
				}
			}
		}
	}

	out, _ := json.MarshalIndent(map[string]interface{}{
		"upstreams": upstreams,
		"servers":   servers,
	}, "", "  ")

	fmt.Fprintln(os.Stdout, string(out))
	os.Exit(0)
}
