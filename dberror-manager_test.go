package jsonapi

import (
	"github.com/kucjac/uni-db"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNew(t *testing.T) {
	Convey("Creating new error handler.", t, func() {
		errorManager := NewDBErrorMgr()

		Convey("The newly created handler use defaultErrorMap by default", func() {
			So(errorManager.dbToRest, ShouldResemble, DefaultErrorMap)

		})
	})
}

func TestLoadCustomErrorMap(t *testing.T) {
	Convey("While having an Error Handler", t, func() {
		errorManager := NewDBErrorMgr()

		Convey("And a prepared custom error map with a custom resterror", func() {
			customError := ErrorObject{Code: "C123", Title: "Custom rest error"}

			customMap := map[unidb.Error]ErrorObject{
				unidb.ErrUnspecifiedError: customError,
			}
			So(customMap, ShouldNotBeNil)

			Convey("For given Error some Error should be returned.", func() {
				someError := unidb.ErrUnspecifiedError.New()
				So(someError, ShouldNotBeNil)
				prevRestErr, err := errorManager.Handle(someError)

				So(err, ShouldBeNil)
				So(prevRestErr, ShouldNotBeNil)

				FocusConvey("While loading custom error map containing given error pair", func() {
					errorManager.LoadCustomErrorMap(customMap)
					So(errorManager.dbToRest, ShouldResemble, customMap)
					So(errorManager, ShouldNotBeNil)

					FocusConvey("Given Error would return different Error", func() {
						afterRestErr, err := errorManager.Handle(someError)
						So(err, ShouldBeNil)
						So(afterRestErr, ShouldNotResemble, prevRestErr)
						So(afterRestErr.ID, ShouldEqual, customError.ID)
					})
				})

			})

		})
	})
}

func TestUpdateErrorEntry(t *testing.T) {
	Convey("Having a ErrorHandler containing default error map", t, func() {
		errorManager := NewDBErrorMgr()

		So(errorManager.dbToRest, ShouldResemble, DefaultErrorMap)

		Convey("Getting a *Error for given Error", func() {
			someErrorProto := unidb.ErrCheckViolation
			someError := someErrorProto.New()

			prevRestErr, err := errorManager.Handle(someError)
			So(err, ShouldBeNil)

			Convey("While using UpdateErrorMapEntry method on the errorManager.", func() {
				customError := ErrorObject{ID: "1234", Title: "My custom Error"}

				errorManager.UpdateErrorEntry(someErrorProto, customError)

				FocusConvey(`Handling again given Error would result
					with different *Error entity`, func() {
					afterRestErr, err := errorManager.Handle(someError)

					So(err, ShouldBeNil)
					So(afterRestErr, ShouldNotResemble, prevRestErr)
					So(afterRestErr.ID, ShouldEqual, customError.ID)
				})

			})
		})
	})
}

func TestHandle(t *testing.T) {
	Convey("Having a ErrorHandler with default error map", t, func() {
		errorManager := NewDBErrorMgr()

		Convey("And a *Error based on the basic Error prototype", func() {
			someError := unidb.ErrUniqueViolation.New()

			Convey(`Then handling given *Error would result
				with some *Error entity`, func() {
				restError, err := errorManager.Handle(someError)

				So(err, ShouldBeNil)
				So(restError, ShouldHaveSameTypeAs, &ErrorObject{})
				So(restError, ShouldNotBeNil)
			})

		})

		Convey("If the *Error is not based on basic Error prototype", func() {
			someCustomError := &unidb.Error{ID: uint(240), Message: "Some error message"}
			Convey("Then handling this error would result with nil *Error and throwing an internal error.", func() {
				restError, err := errorManager.Handle(someCustomError)

				So(err, ShouldBeError)
				So(restError, ShouldBeNil)
			})
		})

		Convey(`If we set a non default error map,
			that may not contain every Error entry as a key`, func() {
			someErrorProto := unidb.ErrSystemError
			customErrorMap := map[unidb.Error]ErrorObject{
				someErrorProto: {ID: "1921", Title: "Some Error"},
			}
			errorManager.LoadCustomErrorMap(customErrorMap)

			Convey(`Then handling a *Error based on the basic Error prototype that is not in
				the custom error map, would throw an internal error
				and a nil *Error.`, func() {

				someDBFromProto := unidb.ErrInvalidSyntax.New()
				otherError, err := errorManager.Handle(someDBFromProto)

				So(err, ShouldBeError)
				So(otherError, ShouldBeNil)

			})

		})

	})

}
