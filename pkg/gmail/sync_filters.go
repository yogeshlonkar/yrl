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
	ID      string   `yaml:"id,omitempty"`
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
		ID:      f.Id,
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
		Id: tempStruct.ID,
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
		log.Fatal().Strs("Ids", []string{f.Id, gf.Id}).Msg("can't sync filters")
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
	log.Trace().Msg("started mergeFilters")
	for _, gf := range existing {
		if matched, ok := filters.lookup(gf); ok {
			matched.sync(gf, push)
			lEvent := log.Debug().Str("ID", matched.Id)
			if matched.Criteria != nil {
				lEvent.Str("From", matched.Criteria.From).Str("Subject", matched.Criteria.Subject)
			}
			switch {
			case matched.pull:
				lEvent.Msg("pull Filter")
				pulls++
			case matched.push:
				lEvent.Msg("push Filter")
				pushes++
			default:
				unchanged++
			}
		} else {
			log.Debug().Str("ID", gf.Id).Msg("unknown Filter found")
			*filters = append(*filters, &Filter{
				gf,
				state{
					pull: true,
				},
			})
			pulls++
		}
	}
	return pulls, pushes, unchanged
}

func deconstruct(field, query string) []string {
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
	log.Trace().Msg("started SyncFilters")
	existing, err := s.svc.ListFilters()
	if err != nil {
		return err
	}
	pulls, pushes, unchanged := mergeFilters(&s.settings.Filters, existing.Filter, s.push)
	var newCount, failed, missing int
	for _, filter := range s.settings.Filters {
		switch {
		case filter.Id == "":
			if s.dryRun {
				log.Info().Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msg("will create filter")
				continue
			}
			gl, err := s.svc.CreateFilter(filter.Filter)
			if err != nil {
				log.Error().Err(err).Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msg("error creating filter")
				failed++
			} else {
				log.Info().Str("ID", gl.Id).Msg("created filter")
				filter.Id = gl.Id
				newCount++
			}
		case filter.push:
			if s.dryRun {
				log.Info().Str("ID", filter.Id).Msg("will update filter")
				continue
			}
			gf, err := s.svc.UpdateFilter(filter.Filter)
			if err != nil {
				log.Error().Err(err).Str("ID", filter.Id).Msg("error updating filter")
				failed++
				pushes--
			} else {
				filter.Filter.Id = gf.Id
			}
		case filter.pull:
			if s.dryRun {
				log.Info().Str("ID", filter.Id).Msg("will pull filter")
				continue
			}
		case !filter.ok:
			log.Error().Str("ID", filter.Id).Str("From", filter.Criteria.From).Str("Subject", filter.Criteria.Subject).Msg("error missing filter")
		}
	}
	if !s.dryRun && (newCount > 0 || pulls > 0 || pushes > 0) {
		s.settings.dirty = true
	}
	lEvent := log.Info().Int("New", newCount).Int("Pull", pulls).Int("Push", pushes).Int("Unchanged", unchanged)
	if s.dryRun {
		lEvent.Msg("after sync filters will be")
	} else {
		lEvent.Int("Failed", failed).Int("Missing", missing).Msg("filters sync complete")
	}
	return nil
}
