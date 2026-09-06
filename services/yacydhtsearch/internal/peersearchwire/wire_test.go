package peersearchwire_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikitakarpei/yacy-rwi-node/yacydhtsearch/internal/peersearchwire"
	"github.com/nikitakarpei/yacy-rwi-node/yacymodel"
	"github.com/nikitakarpei/yacy-rwi-node/yacyproto"
)

const responseLimit = 1 << 20

type recordedOutcome struct {
	answered    int
	refused     int
	unreachable int
	unreadable  int
}

func (r *recordedOutcome) PeerAnswered(context.Context, string, int)           { r.answered++ }
func (r *recordedOutcome) PeerRefused(context.Context, string, int)            { r.refused++ }
func (r *recordedOutcome) PeerUnreachable(context.Context, string, error)      { r.unreachable++ }
func (r *recordedOutcome) PeerAnswerUnreadable(context.Context, string, error) { r.unreadable++ }

func peerAnswering(t *testing.T, body string, status int) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(status)
			_, _ = writer.Write([]byte(body))
		},
	))
	t.Cleanup(server.Close)

	return server.URL
}

func searchAnswerHolding(address string) string {
	return yacyproto.SearchResponse{
		Count: 1,
		Resources: []yacyproto.SearchResource{
			{Metadata: yacymodel.URLMetadata{Address: address, Title: "Weather"}},
		},
	}.Encode().Encode()
}

func TestAPeerAnswerBecomesResultItems(t *testing.T) {
	t.Parallel()

	observer := &recordedOutcome{}
	wire := peersearchwire.New(
		http.DefaultClient,
		responseLimit,
		peersearchwire.PeerSearchObservers{observer},
	)

	items := wire.Search(
		t.Context(),
		peerAnswering(t, searchAnswerHolding("https://example.org/weather"), http.StatusOK),
		yacyproto.SearchRequest{NetworkName: "freeworld"},
	)

	if len(items) != 1 || items[0].Address != "https://example.org/weather" {
		t.Fatalf("Search = %+v, want the address the peer reported", items)
	}
	if observer.answered != 1 {
		t.Fatalf("PeerAnswered reported %d times, want once", observer.answered)
	}
}

func TestAPeerThatRefusesTheSearchYieldsNoItems(t *testing.T) {
	t.Parallel()

	observer := &recordedOutcome{}
	wire := peersearchwire.New(http.DefaultClient, responseLimit, observer)

	items := wire.Search(
		t.Context(),
		peerAnswering(t, "", http.StatusServiceUnavailable),
		yacyproto.SearchRequest{},
	)

	if len(items) != 0 || observer.refused != 1 {
		t.Fatalf("Search = %+v with %d refusals, want none and one", items, observer.refused)
	}
}

func TestAPeerThatCannotBeReachedYieldsNoItems(t *testing.T) {
	t.Parallel()

	observer := &recordedOutcome{}
	wire := peersearchwire.New(http.DefaultClient, responseLimit, observer)

	items := wire.Search(t.Context(), "http://127.0.0.1:1", yacyproto.SearchRequest{})

	if len(items) != 0 || observer.unreachable != 1 {
		t.Fatalf("Search = %+v with %d unreachable, want none and one", items, observer.unreachable)
	}
}
