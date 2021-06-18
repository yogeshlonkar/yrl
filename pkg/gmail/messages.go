package gmail

import (
	"bytes"
	"errors"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/gmail/v1"
)

func (g gmailService) ListMessages(q, pageToken string, maxResults int) (*gmail.ListMessagesResponse, error) {
	call := g.srv.Users.Messages.List(g.user)
	if q != "" {
		call.Q(q)
	}
	if pageToken != "" {
		call.PageToken(pageToken)
	}

	if maxResults > 0 {
		call.MaxResults(int64(maxResults))
	} else {
		call.MaxResults(10000)
	}
	return call.Do()
}

func (g gmailService) FindMessagesAndApplyLabel(from, subject string, labelIds []string) error {
	ids := make([]string, 0)
	var buf bytes.Buffer
	if from != "" {
		buf.WriteString("from:(" + from + ")")
	}
	if subject != "" {
		if buf.Len() > 0 {
			buf.WriteString(" OR ")
		}
		buf.WriteString("subject:(" + subject + ")")
	}
	if buf.Len() == 0 {
		return errors.New("No query to find messages")
	}
	q := buf.String()
	resp, err := g.ListMessages(q, "", -1)
	for {
		if err != nil {
			return err
		}
		for _, msg := range resp.Messages {
			ids = append(ids, msg.Id)
		}
		if resp.NextPageToken == "" {
			break
		}
		resp, err = g.ListMessages(q, resp.NextPageToken, -1)
	}
	if len(ids) == 0 {
		log.Warn().Strs("Labels", labelIds).Str("From", from).Str("Subject", subject).Msg("No message found to apply label")
		return nil
	} else {
		log.Debug().Int("Message", len(ids)).Msg("Applied label")
	}
	return g.ApplyLabelToMessages(labelIds, ids)
}

func (g gmailService) ApplyLabelToMessages(labelIds, ids []string) error {
	lastIndex := 999
	for len(ids) > lastIndex {
		bIds := ids[:lastIndex]
		batch := &gmail.BatchModifyMessagesRequest{
			Ids:         bIds,
			AddLabelIds: labelIds,
		}
		if err := g.srv.Users.Messages.BatchModify(g.user, batch).Do(); err != nil {
			return err
		}
		ids = ids[lastIndex:]
	}
	batch := &gmail.BatchModifyMessagesRequest{
		Ids:         ids,
		AddLabelIds: labelIds,
	}
	return g.srv.Users.Messages.BatchModify(g.user, batch).Do()
}
