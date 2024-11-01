package models

import (
	"time"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/router/routes/models/shared"
)

type TemplateDataIndex struct {
	JSCachePurgeTimestamp int64

	shared.TemplateDataLettr
}

func (tdi TemplateDataIndex) New(l language.Language, p puzzle.Puzzle, pastWords []puzzle.Word, imprintUrl string, revision string, faviconPath string) TemplateDataIndex {
	tdi.TemplateDataLettr = tdi.TemplateDataLettr.New(l, p, pastWords, imprintUrl, revision, faviconPath)
	tdi.JSCachePurgeTimestamp = time.Now().Unix()

	return tdi
}
