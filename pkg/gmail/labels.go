package gmail

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

func (g gmailService) ListLabels() (*gmail.ListLabelsResponse, error) {
	return g.srv.Users.Labels.List(g.user).Do()
}

func (g gmailService) GetLabel(id string) (*gmail.Label, error) {
	return g.srv.Users.Labels.Get(g.user, id).Do()
}

func (g gmailService) CreateLabel(l *gmail.Label) (*gmail.Label, error) {
	return g.srv.Users.Labels.Create(g.user, l).Do()
}

func (g gmailService) UpdateLabel(l *gmail.Label) (*gmail.Label, error) {
	return g.srv.Users.Labels.Update(g.user, l.Id, l).Do()
}

func (g gmailService) GetLabelID(invalidate bool, labelName string) (string, error) {
	if invalidate || len(labelCache) == 0 {
		labelResp, err := g.ListLabels()
		if err != nil {
			return "", err
		}
		labelCache = labelResp.Labels
	}
	for _, label := range labelCache {
		if label.Name == labelName {
			return label.Id, nil
		}
	}
	return "", fmt.Errorf("no id found for label %s", labelName)
}
