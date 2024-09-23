package l10n

import (
	"encoding/json"
	"fmt"

	"github.com/Nidal-Bakir/go-todo-backend/internal/utils"
	"github.com/rs/zerolog"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	bundle    *i18n.Bundle
	locales   = map[string]*Localizer{}
	languages = []string{}
)

type Localizer struct {
	l      *i18n.Localizer
	logger zerolog.Logger
}

func InitL10n(path string, langs []string, logger zerolog.Logger) {
	utils.Assert(len(langs) != 0, "The langs slice can not be empty")
	languages = langs

	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	logEvent := logger.Debug()
	for _, lang := range languages {
		filePath := fmt.Sprintf(path+"/%s.json", lang)
		bundle.MustLoadMessageFile(filePath)
		locales[lang] = &Localizer{l: i18n.NewLocalizer(bundle, lang), logger: logger}

		logEvent.Str(lang, filePath)
	}
	logEvent.Msg("Localiztion files loaded")
}

func GetLocalizer(lang string) *Localizer {
	if _, ok := locales[lang]; !ok {
		l := locales[languages[0]]
		l.logger.Error().Msgf("Language %s not found, will default to %s", lang, languages[0])
		return l
	}
	return locales[lang]
}

func (l *Localizer) GetWithId(id string) string {
	return l.localizeMsg(id, nil, nil)
}

func (l *Localizer) GetWithPluralCount(id string, pluralCount int) string {
	return l.localizeMsg(id, nil, pluralCount)
}

func (l *Localizer) GetWithData(id string, data map[string]interface{}) string {
	utils.AssertDev(data != nil, "The data map can not be nil")
	utils.AssertDev(len(data) != 0, "The data map should not be empty")

	return l.localizeMsg(id, data, nil)
}

func (l *Localizer) Get(id string, data map[string]string, pluralCount int) string {
	return l.localizeMsg(id, data, pluralCount)
}

func (l *Localizer) localizeMsg(id string, data interface{}, pluralCount interface{}) string {
	cfg := &i18n.LocalizeConfig{
		DefaultMessage: defaultMessage(id),
		TemplateData:   data,
		PluralCount:    pluralCount,
	}

	str, err := l.l.Localize(cfg)
	if err != nil {
		errLog := l.logger.Error().Err(err).Str("id", id)
		if d, ok := data.(map[string]interface{}); ok {
			errLog.Fields(d)
		}
		if pluralCount != nil {
			errLog.Any("pluralCount", pluralCount)
		}
		errLog.Msg("Error getting localized message")

		str = id
	}

	return str
}

func defaultMessage(id string) *i18n.Message {
	return &i18n.Message{
		ID:    id,
		Other: id,
		Zero:  id,
		One:   id,
		Two:   id,
		Few:   id,
		Many:  id,
	}
}
