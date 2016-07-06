package manifest

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"

	"github.com/InnovaCo/serve/manifest/loader"
	"github.com/InnovaCo/serve/manifest/processor"
)

var varsFilterRegexp = regexp.MustCompile("[^A-z0-9_\\.]")

func Load(path string, vars map[string]string) *Manifest {
	tree, err := loader.LoadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	for k, v := range vars {
		tree.Set(v, "vars", varsFilterRegexp.ReplaceAllString(k, "_"))
	}

	for name, proc := range processor.GetAll() {
		tree, err = proc.Process(tree)
		if err != nil {
			log.Fatalf("Error in processor '%s': %v", name, err)
		}
	}

	log.Println("\n", tree.StringIndent("", "  "))

	return &Manifest{tree: tree}
}

type Manifest struct {
	tree *gabs.Container
}

func (m Manifest) String() string {
	return m.tree.String()
}

func (m Manifest) GetString(path string) string {
	return fmt.Sprintf("%v", m.tree.Path(path).Data())
}

func (m Manifest) FindPlugins(plugin string) ([]PluginPair, error) {
	plugin = varsFilterRegexp.ReplaceAllString(plugin, "_")

	tree := m.tree.Path(plugin)
	result := make([]PluginPair, 0)

	if _, ok := tree.Data().([]interface{}); ok {
		arr, _ := tree.Children()
		for _, item := range arr {
			if _, ok := item.Data().(string); ok {
				result = append(result, makePluginPair(plugin, item))
			} else if res, err := item.ChildrenMap(); err == nil {
				for subplugin, data := range res {
					result = append(result, makePluginPair(plugin+"."+subplugin, data))
					break
				}
			}
		}
	} else {
		if tree.Data() == nil {
			tree = m.tree.Path("vars")
		}

		result = append(result, makePluginPair(plugin, tree))
	}

	return result, nil
}

func makePluginPair(plugin string, data *gabs.Container) PluginPair {
	if s, ok := data.Data().(string); ok {
		obj := gabs.New()
		ns := strings.Split(plugin, ".")
		obj.Set(s, ns[len(ns)-1])
		data = obj
	}

	return PluginPair{plugin, PluginRegestry.Get(plugin), Manifest{data}}
}
