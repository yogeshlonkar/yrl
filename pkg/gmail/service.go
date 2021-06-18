package gmail

import (
	"context"
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Service interface {
	LabelService
	FilterService
	MessageService
}

type LabelService interface {
	CreateLabel(l *gmail.Label) (*gmail.Label, error)
	GetLabel(id string) (*gmail.Label, error)
	ListLabels() (*gmail.ListLabelsResponse, error)
	UpdateLabel(l *gmail.Label) (*gmail.Label, error)
	GetLabelID(invalidate bool, labelName string) (string, error)
}

type FilterService interface {
	CreateFilter(f *gmail.Filter) (*gmail.Filter, error)
	ListFilters() (*gmail.ListFiltersResponse, error)
	UpdateFilter(f *gmail.Filter) (*gmail.Filter, error)
}

type MessageService interface {
	ListMessages(q, pageToken string, maxResults int) (*gmail.ListMessagesResponse, error)
	ApplyLabelToMessages(labelIds, ids []string) error
}

type gmailService struct {
	user string
	srv  *gmail.Service
}

func NewService(ctx context.Context, creds, user string) Service {
	b, err := ioutil.ReadFile(creds)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read client secret file")
	}
	config, err := google.JWTConfigFromJSON(b, gmail.GmailModifyScope, gmail.GmailSettingsBasicScope)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get jwt")
	}
	config.Subject = user
	tokenSource := config.TokenSource(ctx)
	srv, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to retrieve Gmail client")
	}
	log.Trace().Str("User", user).Msg("created gmailService")
	return &gmailService{
		user,
		srv,
	}
}
