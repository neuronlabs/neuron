package jsonapi

import (
	"github.com/kucjac/uni-logger"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetLanguage(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)
	req.Header.Add(headerAcceptLanguage, "pl")
	h := prepareHandler([]language.Tag{language.Polish, language.BritishEnglish})

	rw := httptest.NewRecorder()

	lang, ok := h.GetLanguage(req, rw)
	assert.True(t, ok)
	assert.Equal(t, language.Polish, lang)

	// No language provided - default set
	req = httptest.NewRequest("GET", "/blogs", nil)
	lang, ok = h.GetLanguage(req, rw)
	assert.True(t, ok)
	assert.Equal(t, language.Polish, lang)

	// Provided invalid language
	req = httptest.NewRequest("GET", "/blogs", nil)
	req.Header.Add(headerAcceptLanguage, "invalid")

	lang, ok = h.GetLanguage(req, rw)
	assert.False(t, ok)
	assert.NotEqual(t, language.Polish, lang)
}

func TestHandlerCheckLanguage(t *testing.T) {
	c := NewController()
	err := c.PrecomputeModels(&ModelI18nSDK{})
	assert.NoError(t, err)

	scope, err := c.NewScope(&ModelI18nSDK{})
	assert.NoError(t, err)
	assert.NotNil(t, scope)

	unilog := unilogger.NewBasicLogger(os.Stderr, "", 0)
	logger := unilogger.MustGetLoggerWrapper(unilog)

	h := NewHandler(c, logger, NewDBErrorMgr())
	h.SetLanguages(language.Polish, language.English)

	// Case 1:
	// Scope with nil value
	rw := httptest.NewRecorder()

	tag, ok := h.CheckValueLanguage(scope, rw)
	assert.False(t, ok)
	assert.NotContains(t, h.SupportedLanguages, tag)
	assert.Equal(t, 500, rw.Result().StatusCode)

	scope.Value = &ModelI18nSDK{ID: 1, Lang: "pl"}

	// Case 2:
	// Scope with some correct value
	rw = httptest.NewRecorder()
	langTag, ok := h.CheckValueLanguage(scope, rw)
	assert.True(t, ok)
	assert.Equal(t, language.Polish, langTag)

	// Case 3:
	// Scope with correct model Value with no language within
	rw = httptest.NewRecorder()
	scope.Value = &ModelI18nSDK{ID: 2}
	_, ok = h.CheckValueLanguage(scope, rw)
	assert.False(t, ok)

	// Case 4:
	// Scope with incorrect language value
	rw = httptest.NewRecorder()
	scope.Value = &ModelI18nSDK{ID: 3, Lang: "invalid_lang"}
	langTag, ok = h.CheckValueLanguage(scope, rw)

	assert.False(t, ok)
	assert.Equal(t, 400, rw.Result().StatusCode)

	// Case 5:
	// Scope with correct language not in the supported languages
	rw = httptest.NewRecorder()
	scope.Value = &ModelI18nSDK{ID: 3, Lang: language.Arabic.String()}
	langTag, ok = h.CheckValueLanguage(scope, rw)

	assert.False(t, ok)
	assert.Equal(t, 406, rw.Result().StatusCode)

	// Case 5:
	// Scope with invalid model
	err = c.PrecomputeModels(&ModelSDK{})
	assert.NoError(t, err)

	scope, err = c.NewScope(&ModelSDK{})
	assert.NoError(t, err)

	scope.Value = &ModelSDK{ID: 1, Name: "Some name"}

	rw = httptest.NewRecorder()
	_, ok = h.CheckValueLanguage(scope, rw)
	assert.False(t, ok)

}
