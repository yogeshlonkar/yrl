package gmail

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/yaml.v3"
)

type Labels []*Label

type Label struct {
	*gmail.Label
	state
}

type labelSetting struct {
	Id     string `yaml:"id,omitempty"`
	Colors string `yaml:"colors,omitempty"`
	Name   string `yaml:"name,omitempty"`
}

func (l *Label) MarshalYAML() (interface{}, error) {
	return &labelSetting{
		Id:     l.Id,
		Name:   l.Name,
		Colors: l.colors(),
	}, nil
}

func (l *Label) UnmarshalYAML(value *yaml.Node) error {
	tempStruct := &labelSetting{}
	if err := value.Decode(&tempStruct); err != nil {
		return err
	}
	l.Label = &gmail.Label{
		Name: tempStruct.Name,
		Id:   tempStruct.Id,
	}
	if tempStruct.Colors != "" {
		if colors := strings.Split(tempStruct.Colors, "/"); len(colors) == 2 {
			l.Label.Color = &gmail.LabelColor{
				BackgroundColor: colors[0],
				TextColor:       colors[1],
			}
		} else {
			return fmt.Errorf("Invalid color %s expected background/text color format example #000000/#ffffff", l.Color)
		}
	}
	return nil
}

func (l *Label) sync(gl *gmail.Label, push bool) {
	if l.Name != gl.Name {
		log.Fatal().Strs("Names", []string{l.Name, gl.Name}).Msgf("Can't sync labels")
	}
	b := &Label{Label: gl}
	if l.equal(b) {
		l.ok = true
	} else {
		if push {
			l.push = true
		} else {
			l.Label = gl
			l.pull = true
		}
	}
}

func (l *Label) equal(b *Label) bool {
	return l.colors() == b.colors()
}

func (l *Label) colors() string {
	if l.Color == nil {
		return ""
	}
	return l.Color.BackgroundColor + "/" + l.Color.TextColor
}

func (labels Labels) lookup(gl *gmail.Label) (*Label, bool) {
	for _, label := range labels {
		if label.Id == gl.Id {
			return label, true
		}
	}
	for _, label := range labels {
		if label.Name == gl.Name {
			return label, true
		}
	}
	return nil, false
}

func mergeLabels(labels *Labels, existing []*gmail.Label, push bool, getLabel func(string) (*gmail.Label, error)) (pulls, pushes, unchanged int) {
	log.Trace().Msg("Started mergeLabels")
	for _, gl := range existing {
		if gl.Type == "user" {
			if matched, ok := labels.lookup(gl); ok {
				gl, err := getLabel(gl.Id)
				if err != nil {
					log.Error().Err(err).Msgf("Couldn't get label %s", gl.Id)
					continue
				}
				matched.sync(gl, push)
				lEvent := log.Debug().Str("Name", matched.Name).Str("Id", matched.Id)
				if matched.pull {
					if gl.Color != nil {
						lEvent.Str("BackgroundColor", gl.Color.BackgroundColor).Str("TextColor", gl.Color.TextColor)
					}
					lEvent.Msg("Pull Label")
					pulls++
				} else if matched.push {
					if matched.Color != nil {
						lEvent.Str("BackgroundColor", matched.Color.BackgroundColor).Str("TextColor", matched.Color.TextColor)
					}
					lEvent.Msg("Push Label")
					pushes++
				} else {
					unchanged++
				}
			} else {
				log.Debug().Str("Name", gl.Name).Str("Id", gl.Id).Msg("Unknown label found")
				*labels = append(*labels, &Label{
					gl,
					state{
						pull: true,
					},
				})
				pulls++
			}
		}
	}
	return
}

func (s *syncService) SyncLabels() error {
	log.Trace().Msg("Started SyncLabels")
	existing, err := s.svc.ListLabels()
	if err != nil {
		return err
	}
	log.Trace().Msg("Go existing labels")
	pulls, pushes, unchanged := mergeLabels(&s.settings.Labels, existing.Labels, s.push, s.svc.GetLabel)
	var new, failed, missing int
	for _, label := range s.settings.Labels {
		if label.Id == "" {
			if s.dryRun {
				log.Info().Str("Name", label.Name).Msg("Will create label")
				continue
			}
			gl, err := s.svc.CreateLabel(label.Label)
			if err != nil {
				log.Error().Err(err).Str("Name", label.Name).Msgf("Error creating label")
				failed++
			} else {
				log.Info().Str("Name", label.Name).Str("Id", gl.Id).Msg("Created label")
				label.Id = gl.Id
				new++
			}
		} else if label.push {
			if s.dryRun {
				log.Info().Str("Id", label.Id).Str("Name", label.Name).Msg("Will update label")
				continue
			}
			_, err := s.svc.UpdateLabel(label.Label)
			if err != nil {
				log.Error().Err(err).Str("Id", label.Id).Msgf("Error updating label")
				failed++
				pushes--
			}
		} else if label.pull {
			if s.dryRun {
				log.Info().Str("Id", label.Id).Str("Name", label.Name).Msg("Will pull label")
				continue
			}
		} else if !label.ok {
			log.Error().Str("Id", label.Id).Str("Name", label.Name).Msgf("Error missing label")
			missing++
		}
	}
	if !s.dryRun && (new > 0 || pulls > 0) {
		s.settings.dirty = true
	}
	lEvent := log.Info().Int("New", new).Int("Pull", pulls).Int("Push", pushes).Int("Unchanged", unchanged)
	if s.dryRun {
		lEvent.Msg("After sync labels will be")
	} else {
		lEvent.Int("Failed", failed).Int("Missing", missing).Msg("Labels sync complete")
	}
	return nil
}
