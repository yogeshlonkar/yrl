package gmail

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/yaml.v3"
)

type Filters []*Filter

type Filter struct {
	*gmail.Filter
	state
}

type filterSetting struct {
	Id      string   `yaml:"id,omitempty"`
	From    []string `yaml:"from,omitempty,flow"`
	Subject []string `yaml:"subject,omitempty,flow"`
	Labels  []string `yaml:"labels,omitempty,flow"`
}

func (f *Filter) MarshalYAML() (interface{}, error) {
	var from, subject, labels []string
	if f.Filter != nil && f.Filter.Criteria != nil {
		from = deconstruct("", f.Filter.Criteria.From)
	}
	if f.Filter != nil && f.Filter.Criteria != nil {
		subject = deconstruct("", f.Filter.Criteria.Subject)
	}
	if f.Filter != nil && f.Action != nil {
		labels = f.Filter.Action.AddLabelIds
	}
	return &filterSetting{
		Id:      f.Id,
		From:    from,
		Subject: subject,
		Labels:  labels,
	}, nil
}

func (f *Filter) UnmarshalYAML(value *yaml.Node) error {
	tempStruct := &filterSetting{}
	if err := value.Decode(&tempStruct); err != nil {
		return err
	}
	f.Filter = &gmail.Filter{
		Id: tempStruct.Id,
		Criteria: &gmail.FilterCriteria{
			From:    construct("", tempStruct.From),
			Subject: construct("", tempStruct.Subject),
		},
		Action: &gmail.FilterAction{
			AddLabelIds: tempStruct.Labels,
		},
	}
	return nil
}

func (f *Filter) sync(gf *gmail.Filter, push bool) {
	if f.Id != gf.Id {
		log.Fatal().Strs("Ids", []string{f.Id, gf.Id}).Msgf("Can't sync filters")
	}
	if f.equal(&Filter{Filter: gf}) {
		f.ok = true
	} else {
		if push {
			f.push = true
		} else {
			f.pull = true
		}
	}
}

func (f *Filter) equal(b *Filter) bool {
	if b == nil || f == nil || f.Filter == nil || b.Filter == nil {
		return false
	}
	if f.Action == nil || b.Action == nil {
		return false
	}
	if len(f.Action.AddLabelIds) != len(b.Action.AddLabelIds) {
		return false
	}
	for i, v := range f.Action.AddLabelIds {
		if v != b.Action.AddLabelIds[i] {
			return false
		}
	}
	if f.Criteria == nil || b.Criteria == nil {
		return false
	}
	if f.Criteria.From != b.Criteria.From {
		return false
	}
	if f.Criteria.Subject != b.Criteria.Subject {
		return false
	}
	return true
}

func (filters Filters) lookup(gf *gmail.Filter) (*Filter, bool) {
	for _, filter := range filters {
		if filter.Id == gf.Id {
			return filter, true
		}
	}
	return nil, false
}

func mergeFilters(filters *Filters, existing []*gmail.Filter, push bool) (pulls, pushes, unchanged int) {
	log.Trace().Msg("Started mergeFilters")
	for _, gf := range existing {
		if matched, ok := filters.lookup(gf); ok {
			matched.sync(gf, push)
			lEvent := log.Debug().Str("Id", matched.Id)
			if matched.Criteria != nil {
				lEvent.Str("From", matched.Criteria.From).Str("Subject", matched.Criteria.Subject)
			}
			if matched.pull {
				lEvent.Msg("Pull Filter")
				pulls++
			} else if matched.push {
				lEvent.Msg("Push Filter")
				pushes++
			} else {
				unchanged++
			}
		} else {
			log.Debug().Str("Id", gf.Id).Msg("Unknown Filter found")
			*filters = append(*filters, &Filter{
				gf,
				state{
					pull: true,
				},
			})
			pulls++
		}
	}
	return
}

func deconstruct(field string, query string) []string {
	if query == "" {
		return nil
	}
	temp := strings.TrimPrefix(query, field+":")
	return strings.Split(temp, " OR ")
}

func construct(field string, query []string) string {
	if len(query) == 0 {
		return ""
	}
	var buf bytes.Buffer
	if field != "" {
		buf.WriteString(field + ":")
	}
	for index, seg := range query {
		if index > 0 {
			buf.WriteString(" OR ")
		}
		buf.WriteString(seg)
	}
	return buf.String()
}

func (s *syncService) SyncFilters() error {
	log.Trace().Msg("Started SyncFilters")
	existing, err := s.svc.ListFilters()
	if err != nil {
		return err
	}
	pulls, pushes, unchanged := mergeFilters(&s.settings.Filters, existing.Filter, s.push)
	var new, failed, missing int
	for _, filter := range s.settings.Filters {
		if filter.Id == "" {
			if s.dryRun {
				log.Info().Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msg("Will create filter")
				continue
			}
			gl, err := s.svc.CreateFilter(filter.Filter)
			if err != nil {
				log.Error().Err(err).Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msgf("Error creating filter")
				failed++
			} else {
				log.Info().Str("Id", gl.Id).Msg("Created filter")
				filter.Id = gl.Id
				new++
			}
		} else if filter.push {
			if s.dryRun {
				log.Info().Str("Id", filter.Id).Msg("Will update filter")
				continue
			}
			gf, err := s.svc.UpdateFilter(filter.Filter)
			if err != nil {
				log.Error().Err(err).Str("Id", filter.Id).Msgf("Error updating filter")
				failed++
				pushes--
			} else {
				filter.Filter.Id = gf.Id
			}
		} else if filter.pull {
			if s.dryRun {
				log.Info().Str("Id", filter.Id).Msg("Will pull filter")
				continue
			}
		} else if !filter.ok {
			log.Error().Str("Id", filter.Id).Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msgf("Error missing filter")
		}
	}
	if !s.dryRun && (new > 0 || pulls > 0 || pushes > 0) {
		s.settings.dirty = true
	}
	lEvent := log.Info().Int("New", new).Int("Pull", pulls).Int("Push", pushes).Int("Unchanged", unchanged)
	if s.dryRun {
		lEvent.Msg("After sync filters will be")
	} else {
		lEvent.Int("Failed", failed).Int("Missing", missing).Msg("Filters sync complete")
	}
	return nil
}
