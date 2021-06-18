package gmail

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Settings struct {
	path    string
	Labels  Labels  `yaml:"labels"`
	Filters Filters `yaml:"filters"`
	dirty   bool
}

type state struct {
	pull, push, ok bool
}

func (s *Settings) update() {
	log.Trace().Msg("started Settings.update")
	file, err := os.Stat(s.path)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading %s", s.path)
	}
	input, err := ioutil.ReadFile(s.path)
	if err != nil {
		log.Fatal().Err(err).Msgf("error reading %s", s.path)
	}
	backup := s.path + "." + time.Now().Format("20060102150405.000")
	err = ioutil.WriteFile(backup, input, file.Mode().Perm())
	if err != nil {
		log.Fatal().Err(err).Msgf("error creating %s", backup)
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling settings")
	}
	node := &yaml.Node{}
	err = yaml.Unmarshal(data, node)
	if err != nil {
		log.Fatal().Err(err).Msg("error Unmarshalling settings in *yaml.Node{}")
	}
	if len(node.Content) == 1 {
		for _, n := range node.Content[0].Content {
			if n.Value == "filters" {
				n.HeadComment = "# These are filters"
			}
		}
	}
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(0)
	err = yamlEncoder.Encode(node)
	if err != nil {
		log.Fatal().Err(err).Msg("error marshaling *yaml.Node{}")
	}
	yamlEncoder.Close()
	err = ioutil.WriteFile(s.path, b.Bytes(), file.Mode().Perm())
	if err != nil {
		log.Fatal().Err(err).Msgf("error updating %s", s.path)
	}
}
