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
	ID     string `yaml:"id,omitempty"`
	Colors string `yaml:"colors,omitempty"`
	Name   string `yaml:"name,omitempty"`
}

func (l *Label) MarshalYAML() (interface{}, error) {
	return &labelSetting{
		ID:     l.Id,
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
		Id:   tempStruct.ID,
	}
	if tempStruct.Colors != "" {
		if colors := strings.Split(tempStruct.Colors, "/"); len(colors) == 2 {
			l.Label.Color = &gmail.LabelColor{
				BackgroundColor: colors[0],
				TextColor:       colors[1],
			}
		} else {
			return fmt.Errorf("invalid color %s expected background/text color format example #000000/#ffffff", l.Color)
		}
	}
	return nil
}

func (l *Label) sync(gl *gmail.Label, push bool) {
	if l.Name != gl.Name {
		log.Fatal().Strs("Names", []string{l.Name, gl.Name}).Msg("can't sync labels")
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
	log.Trace().Msg("started mergeLabels")
	for _, gl := range existing {
		if gl.Type == "user" {
			if matched, ok := labels.lookup(gl); ok {
				gl, err := getLabel(gl.Id)
				if err != nil {
					log.Error().Err(err).Msgf("couldn't get label %s", gl.Id)
					continue
				}
				matched.sync(gl, push)
				lEvent := log.Debug().Str("Name", matched.Name).Str("ID", matched.Id)
				switch {
				case matched.pull:
					if gl.Color != nil {
						lEvent.Str("BackgroundColor", gl.Color.BackgroundColor).Str("TextColor", gl.Color.TextColor)
					}
					lEvent.Msg("pull Label")
					pulls++
				case matched.push:
					if matched.Color != nil {
						lEvent.Str("BackgroundColor", matched.Color.BackgroundColor).Str("TextColor", matched.Color.TextColor)
					}
					lEvent.Msg("push Label")
					pushes++
				default:
					unchanged++
				}
			} else {
				log.Debug().Str("Name", gl.Name).Str("ID", gl.Id).Msg("unknown label found")
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
	return pulls, pushes, unchanged
}

func (s *syncService) SyncLabels() error {
	log.Trace().Msg("started SyncLabels")
	existing, err := s.svc.ListLabels()
	if err != nil {
		return err
	}
	log.Trace().Msg("go existing labels")
	pulls, pushes, unchanged := mergeLabels(&s.settings.Labels, existing.Labels, s.push, s.svc.GetLabel)
	var newCount, failed, missing int
	for _, label := range s.settings.Labels {
		switch {
		case label.Id == "":
			if s.dryRun {
				log.Info().Str("Name", label.Name).Msg("will create label")
				continue
			}
			gl, err := s.svc.CreateLabel(label.Label)
			if err != nil {
				log.Error().Err(err).Str("Name", label.Name).Msg("error creating label")
				failed++
			} else {
				log.Info().Str("Name", label.Name).Str("ID", gl.Id).Msg("created label")
				label.Id = gl.Id
				newCount++
			}
		case label.push:
			if s.dryRun {
				log.Info().Str("ID", label.Id).Str("Name", label.Name).Msg("will update label")
				continue
			}
			_, err := s.svc.UpdateLabel(label.Label)
			if err != nil {
				log.Error().Err(err).Str("ID", label.Id).Msg("error updating label")
				failed++
				pushes--
			}
		case label.pull:
			if s.dryRun {
				log.Info().Str("ID", label.Id).Str("Name", label.Name).Msg("will pull label")
				continue
			}
		case !label.ok:
			log.Error().Str("ID", label.Id).Str("Name", label.Name).Msg("error missing label")
			missing++
		}
	}
	if !s.dryRun && (newCount > 0 || pulls > 0) {
		s.settings.dirty = true
	}
	lEvent := log.Info().Int("New", newCount).Int("Pull", pulls).Int("Push", pushes).Int("Unchanged", unchanged)
	if s.dryRun {
		lEvent.Msg("after sync labels will be")
	} else {
		lEvent.Int("Failed", failed).Int("Missing", missing).Msg("labels sync complete")
	}
	return nil
}
