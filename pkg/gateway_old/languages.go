package gateway

import (
	"fmt"
	"golang.org/x/text/language"
	display "golang.org/x/text/language/display"
	"net/http"
	"strings"
)

func buildLanguageTags(languages string) ([]language.Tag, []*ErrorObject) {
	var (
		tags []language.Tag
		errs []*ErrorObject
	)

	splitted := strings.SplitN(languages, annotationSeperator, 5)
	for _, lang := range splitted {
		tag, err := language.Parse(lang)
		if err != nil {
			switch v := err.(type) {
			case language.ValueError:
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided language: '%s' contains unknown value: '%s'.", lang, v.Subtag())
				errs = append(errs, errObj)
			default:
				// syntax error
				errObj := ErrInvalidQueryParameter.Copy()
				errObj.Detail = fmt.Sprintf("Provided language: '%s' is not syntetically correct.", lang)
				errs = append(errs, errObj)
			}
			continue
		}
		if len(errs) == 0 {
			tags = append(tags, tag)
		}
	}
	return tags, errs
}

func (h *Handler) GetLanguage(
	req *http.Request,
	rw http.ResponseWriter,
) (tag language.Tag, ok bool) {
	tags, _, err := language.ParseAcceptLanguage(req.Header.Get(headerAcceptLanguage))
	if err != nil {
		errObj := ErrInvalidHeaderValue.Copy()
		errObj.Detail = err.Error()
		h.MarshalErrors(rw, errObj)
		return
	}
	tag, _, _ = h.LanguageMatcher.Match(tags...)
	rw.Header().Add(headerContentLanguage, tag.String())
	return tag, true
}

func (h *Handler) DisplaySupportedLanguages() []string {
	namer := display.Tags(language.English)
	var names []string = make([]string, len(h.SupportedLanguages))
	for i, lang := range h.SupportedLanguages {
		names[i] = namer.Name(lang)
	}
	return names
}

// CheckLanguage checks the language value within given scope's Value.
// It parses the field's value into language.Tag, and if no matching language is provided or the // language is not supported, it writes error to the response.
func (h *Handler) CheckValueLanguage(
	scope *Scope,
	rw http.ResponseWriter,
) (langtag language.Tag, ok bool) {
	lang, err := scope.GetLangtagValue()
	if err != nil {
		h.log.Errorf("Error while getting langtag from scope: '%v', Error: %v", scope.Struct.GetType(), err)
		h.MarshalInternalError(rw)
		return
	}

	if lang == "" {
		errObj := ErrInvalidInput.Copy()
		errObj.Detail = fmt.Sprintf("Provided object with no language tag required. Supported languages: %v.", strings.Join(h.DisplaySupportedLanguages(), ","))
		h.MarshalErrors(rw, errObj)
		return
	} else {
		// Parse the language value from taken from the scope's value
		parseTag, err := language.Parse(lang)
		if err != nil {
			// if the error is a value error, then the subtag is invalid
			// this way the langtag should be a well-formed base tag
			if _, isValErr := err.(language.ValueError); !isValErr {
				// If provided language tag is not valid return a user error
				errObj := ErrInvalidInput.Copy()
				errObj.Detail = fmt.
					Sprintf("Provided invalid language tag: '%v'. Error: %v", lang, err)
				h.MarshalErrors(rw, errObj)
				return
			}
		}
		var confidence language.Confidence
		langtag, _, confidence = h.Controller.Matcher.Match(parseTag)
		// If the confidence is low or none send unsupported.
		if confidence <= language.Low {
			// language not supported
			errObj := ErrLanguageNotAcceptable.Copy()
			errObj.Detail = fmt.Sprintf("The language: '%s' is not supported. This document supports following languages: %v",
				lang,
				strings.Join(h.Controller.displaySupportedLanguages(), ","))
			h.MarshalErrors(rw, errObj)
			return
		}
	}
	b, _ := langtag.Base()
	err = scope.SetLangtagValue(b.String())
	if err != nil {
		h.log.Error(err)
		h.MarshalInternalError(rw)
		return
	}
	ok = true
	return
}

// HeaderContentLanguage sets the response Header 'Content-Language' to the lang tag provided in
// argument.
func (h *Handler) HeaderContentLanguage(rw http.ResponseWriter, langtag language.Tag) {
	rw.Header().Add(headerAcceptLanguage, langtag.String())
}
