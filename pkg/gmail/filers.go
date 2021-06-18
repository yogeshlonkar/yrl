package gmail

import (
	"strings"

	"google.golang.org/api/gmail/v1"
)

var (
	labelCache []*gmail.Label
)

func (g gmailService) CreateFilter(f *gmail.Filter) (*gmail.Filter, error) {
	if err := g.mapLabelIds(f); err != nil {
		return nil, err
	}
	if err := g.FindMessagesAndApplyLabel(f.Criteria.From, f.Criteria.Subject, f.Action.AddLabelIds); err != nil {
		return nil, err
	}
	return g.srv.Users.Settings.Filters.Create(g.user, f).Do()
}

func (g gmailService) ListFilters() (*gmail.ListFiltersResponse, error) {
	return g.srv.Users.Settings.Filters.List(g.user).Do()
}

func (g gmailService) UpdateFilter(f *gmail.Filter) (*gmail.Filter, error) {
	if err := g.mapLabelIds(f); err != nil {
		return nil, err
	}
	if err := g.FindMessagesAndApplyLabel(f.Criteria.From, f.Criteria.Subject, f.Action.AddLabelIds); err != nil {
		return nil, err
	}
	if err := g.srv.Users.Settings.Filters.Delete(g.user, f.Id).Do(); err != nil {
		return nil, err
	}
	f.Id = ""
	return g.srv.Users.Settings.Filters.Create(g.user, f).Do()
}

func (g gmailService) mapLabelIds(f *gmail.Filter) error {
	for i, label := range f.Action.AddLabelIds {
		if !strings.HasPrefix(label, "Label_") {
			if id, err := g.GetLabelById(false, label); err == nil {
				f.Action.AddLabelIds[i] = id
			} else {
				return err
			}
		}
	}
	return nil
}
