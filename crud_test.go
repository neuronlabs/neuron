package jsonapi

import (
	"bytes"
	"flag"
	"github.com/kucjac/uni-logger"
	"golang.org/x/text/language"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
)

var debugFlag *bool = flag.Bool("debug", false, "Log level debug")

type funcScopeMatcher func(*Scope) bool

var (
	defaultLanguages = []language.Tag{language.English, language.Polish}
	blogModels       = []interface{}{&BlogSDK{}, &PostSDK{}, &CommentSDK{}, &AuthorSDK{}}
)

func getHttpPair(method, target string, body io.Reader,
) (rw *httptest.ResponseRecorder, req *http.Request) {
	req = httptest.NewRequest(method, target, body)
	req.Header.Add("Content-Type", MediaType)
	rw = httptest.NewRecorder()
	return
}

func prepareModelHandlers(models ...interface{}) (handlers []*ModelHandler) {
	for _, model := range models {
		handler, err := NewModelHandler(model, nil, []EndpointType{Get, GetRelated, GetRelationship, List, Create, Patch, Delete}...)
		if err != nil {
			panic(err)
		}
		handlers = append(handlers, handler)
	}
	return
}

func prepareHandler(languages []language.Tag, models ...interface{}) *Handler {

	c := DefaultController(func() []language.Tag { return languages })

	basic := unilogger.NewBasicLogger(os.Stdout, "", log.Ldate|log.Lshortfile|log.Ltime)
	if *debugFlag {
		basic.SetLevel(unilogger.DEBUG)
		// basic.Info("LogLevel: DEBUG")
	} else {
		basic.SetLevel(unilogger.INFO)
		// basic.Info("LogLevel: INFO")
	}

	c.logger = basic

	// logger := unilogger.MustGetLoggerWrapper(basic)

	h := NewHandler(c, basic, NewDBErrorMgr())
	err := c.PrecomputeModels(models...)
	if err != nil {
		panic(err)
	}

	h.AddModelHandlers(prepareModelHandlers(models...)...)

	h.SetLanguages(languages...)

	return h
}

func matchScopeByType(model interface{}) funcScopeMatcher {
	return func(scope *Scope) bool {
		return isSameType(model, scope)
	}
}

func (h *Handler) getModelJSON(
	model interface{},
) *bytes.Buffer {
	scope, err := h.Controller.NewScope(model)
	if err != nil {
		panic(err)
	}

	scope.Value = model
	h.log.Debugf("getModelJSON Value: %#v", scope.Value)
	payload, err := h.Controller.MarshalScope(scope)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	h.log.Debugf("Payload: %#v", payload.(*OnePayload).Data)
	if err = MarshalPayload(buf, payload); err != nil {
		panic(err)
	}
	return buf

}

func matchScopeByTypeAndID(model interface{}, id interface{}) funcScopeMatcher {
	return func(scope *Scope) bool {
		if matched := isSameType(model, scope); !matched {
			return false
		}

		if scope.Value == nil {
			return false
		}

		v := reflect.ValueOf(scope.Value)
		if v.Type().Kind() != reflect.Ptr {
			return false
		}

		idIndex := scope.Struct.GetPrimaryField().GetFieldIndex()
		return reflect.DeepEqual(id, v.Elem().Field(idIndex).Interface())
	}
}

func isSameType(model interface{}, scope *Scope) bool {
	return reflect.TypeOf(model) == scope.Struct.GetType()
}
