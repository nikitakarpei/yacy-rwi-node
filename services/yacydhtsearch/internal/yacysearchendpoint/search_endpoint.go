// Package yacysearchendpoint serves YaCy's public /yacysearch.json contract.
package yacysearchendpoint

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchquery"
	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/searchresult"
)

const (
	Path = "/yacysearch.json"

	fieldQuery          = "query"
	fieldStartRecord    = "startRecord"
	fieldMaximumRecords = "maximumRecords"
	fieldLanguage       = "lr"

	defaultMaximumRecords = 10
)

type QueryRankings interface {
	RankingFor(ctx context.Context, query searchquery.Query) searchresult.Ranking
}

type SearchEndpoint struct {
	rankings      QueryRankings
	recordCeiling int
}

func NewMux(rankings QueryRankings, recordCeiling int) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(Path, New(rankings, recordCeiling))

	return mux
}

func New(rankings QueryRankings, recordCeiling int) SearchEndpoint {
	return SearchEndpoint{rankings: rankings, recordCeiling: recordCeiling}
}

func (e SearchEndpoint) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	form := request.URL.Query()

	ranking := e.rankings.RankingFor(request.Context(), queryOf(form))
	page := ranking.PageFrom(startRecordOf(form), e.maximumRecordsOf(form))

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(writer).Encode(searchPageFrom(page))
}

func queryOf(form url.Values) searchquery.Query {
	query := searchquery.QueryFrom(form.Get(fieldQuery))
	query.Language = form.Get(fieldLanguage)

	return query
}

func startRecordOf(form url.Values) int {
	return max(0, intOf(form.Get(fieldStartRecord), 0))
}

func (e SearchEndpoint) maximumRecordsOf(form url.Values) int {
	return min(
		e.recordCeiling,
		max(1, intOf(form.Get(fieldMaximumRecords), defaultMaximumRecords)),
	)
}

func intOf(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}
