package generate_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"hop.top/uri/handle/generate"
)

type handlerFixture struct {
	Cases []struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Spec     struct {
			Vendor      string `json:"vendor"`
			App         string `json:"app"`
			Instance    string `json:"instance"`
			Language    string `json:"language"`
			Scheme      string `json:"scheme"`
			Version     string `json:"version"`
			Channel     string `json:"channel"`
			AppPath     string `json:"appPath"`
			DisplayName string `json:"displayName"`
		} `json:"spec"`
		Expected struct {
			HandlerID        string   `json:"handlerId"`
			DesktopFilename  string   `json:"desktopFilename"`
			RenderedContains []string `json:"renderedContains"`
		} `json:"expected"`
	} `json:"cases"`
	InvalidCases []struct {
		Name string `json:"name"`
		Spec struct {
			Vendor      string `json:"vendor"`
			App         string `json:"app"`
			Instance    string `json:"instance"`
			Language    string `json:"language"`
			Scheme      string `json:"scheme"`
			Version     string `json:"version"`
			Channel     string `json:"channel"`
			AppPath     string `json:"appPath"`
			DisplayName string `json:"displayName"`
		} `json:"spec"`
	} `json:"invalidCases"`
}

func loadHandlerFixture(t *testing.T) handlerFixture {
	t.Helper()

	raw, err := os.ReadFile("../../../spec/fixtures/handler-contract.json")
	require.NoError(t, err)

	var fixture handlerFixture
	require.NoError(t, json.Unmarshal(raw, &fixture))
	return fixture
}

func specFromFixture(s struct {
	Vendor      string `json:"vendor"`
	App         string `json:"app"`
	Instance    string `json:"instance"`
	Language    string `json:"language"`
	Scheme      string `json:"scheme"`
	Version     string `json:"version"`
	Channel     string `json:"channel"`
	AppPath     string `json:"appPath"`
	DisplayName string `json:"displayName"`
}) generate.HandlerSpec {
	return generate.HandlerSpec{
		Vendor:      s.Vendor,
		App:         s.App,
		Instance:    s.Instance,
		Language:    generate.Language(s.Language),
		Scheme:      s.Scheme,
		Version:     s.Version,
		Channel:     s.Channel,
		AppPath:     s.AppPath,
		DisplayName: s.DisplayName,
	}
}

func testSpec() generate.HandlerSpec {
	return generate.HandlerSpec{
		Vendor:      "hop-top",
		App:         "scheme",
		Language:    generate.LanguageGo,
		Scheme:      "task",
		Version:     "0.2.0-alpha.0",
		Channel:     "alpha",
		AppPath:     "/usr/bin/task",
		DisplayName: "Hop Task Handler",
	}
}

func TestHandlerSpec_HandlerID(t *testing.T) {
	fixture := loadHandlerFixture(t)

	id, err := specFromFixture(fixture.Cases[0].Spec).HandlerID()
	require.NoError(t, err)
	assert.Equal(t, fixture.Cases[0].Expected.HandlerID, id)

	spec := specFromFixture(fixture.Cases[1].Spec)
	id, err = spec.HandlerID()
	require.NoError(t, err)
	assert.Equal(t, fixture.Cases[1].Expected.HandlerID, id)
}

func TestHandlerSpec_Validate(t *testing.T) {
	fixture := loadHandlerFixture(t)
	for _, tt := range fixture.InvalidCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert.Error(t, specFromFixture(tt.Spec).Validate())
		})
	}
}

func TestSnippet_Platforms(t *testing.T) {
	fixture := loadHandlerFixture(t)
	for _, tt := range fixture.Cases {
		t.Run(tt.Name, func(t *testing.T) {
			out, err := generate.Snippet(tt.Platform, specFromFixture(tt.Spec))
			require.NoError(t, err)
			for _, expected := range tt.Expected.RenderedContains {
				assert.Contains(t, out, expected)
			}
		})
	}
}

func TestSnippet_UnknownPlatform(t *testing.T) {
	_, err := generate.Snippet("amiga", testSpec())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown platform")
}
