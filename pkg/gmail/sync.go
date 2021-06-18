package gmail

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type SyncService interface {
	SyncLabels() error
	SyncFilters() error
	ExpectedSettings() *Settings
}

type syncService struct {
	settings     *Settings
	svc          Service
	dryRun, push bool
}

func NewSyncService(file string, svc Service, dryRun, push bool) (syncSvc SyncService, closeSvc func()) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to read settings %s file", file)
	}
	settings := &Settings{
		path: file,
	}
	if err := yaml.Unmarshal(b, settings); err != nil {
		log.Fatal().Err(err).Msg("unable to parse settings")
	}
	log.Trace().Str("Path", settings.path).Msg("created syncService")
	syncSvc = &syncService{
		settings,
		svc,
		dryRun,
		push,
	}
	closeSvc = func() {
		if settings.dirty {
			settings.update()
		}
	}
	return
}

func (s *syncService) ExpectedSettings() *Settings {
	return s.settings
}
