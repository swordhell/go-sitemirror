package internal_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "go-sitemirror/web/internal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServeInfo", func() {

	newServeInfo := func() (ServeInfo, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		si := NewServeInfo(false, w)

		return si, w
	}

	Context("StatusCode", func() {
		It("should return status code", func() {
			statusCode := http.StatusOK

			si, _ := newServeInfo()
			si.SetStatusCode(statusCode)

			Expect(si.GetStatusCode()).To(Equal(statusCode))
		})

		It("should reset error on new status code", func() {
			si, _ := newServeInfo()
			si.OnCacheNotFound(errors.New("text"))
			Expect(si.HasError()).To(BeTrue())

			si.SetStatusCode(http.StatusOK)
			Expect(si.HasError()).To(BeFalse())
		})
	})

	It("should return content info", func() {
		body := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
		si, _ := newServeInfo()
		si.WriteBody(body)

		cl, cw := si.GetContentInfo()
		Expect(cl).To(Equal(int64(len(body))))
		Expect(cw).To(Equal(cl))
	})

	Describe("getExpires", func() {
		It("should return expires", func() {
			expires := time.Now()

			si, _ := newServeInfo()
			si.SetExpires(expires)

			Expect(*si.GetExpires()).To(Equal(expires))
		})

		It("should return no expires", func() {
			si, _ := newServeInfo()

			Expect(si.GetExpires()).To(BeNil())
		})
	})

	Describe("Error", func() {
		It("should return error", func() {
			si, _ := newServeInfo()
			si.OnCacheNotFound(errors.New("text"))

			Expect(si.HasError()).To(BeTrue())

			t, e := si.GetError()
			Expect(t).To(BeNumerically(">", 0))
			Expect(e).To(HaveOccurred())
		})

		It("should return no error", func() {
			si, _ := newServeInfo()

			Expect(si.HasError()).To(BeFalse())
		})
	})

	Describe("OnXXX", func() {
		It("should handle method not allowed", func() {
			si, _ := newServeInfo()
			si.OnMethodNotAllowed()

			Expect(si.GetStatusCode()).To(BeNumerically(">=", 400))
		})

		It("should handle cache not found", func() {
			si, _ := newServeInfo()
			si.OnCacheNotFound(errors.New("text"))

			Expect(si.GetStatusCode()).To(BeNumerically(">=", 400))
			Expect(si.HasError()).To(BeTrue())
		})

		It("should handle no status code", func() {
			si, _ := newServeInfo()
			si.OnNoStatusCode(ErrorCacheNotFound, "Error message")

			Expect(si.GetStatusCode()).To(BeNumerically(">=", 500))
			Expect(si.HasError()).To(BeTrue())
		})

		It("should handle broken header", func() {
			si, _ := newServeInfo()
			si.OnBrokenHeader(ErrorCacheNotFound, "Error message")

			Expect(si.GetStatusCode()).To(BeNumerically(">=", 500))
			Expect(si.HasError()).To(BeTrue())
		})

		It("should handle cross host invalid path", func() {
			si, _ := newServeInfo()
			si.OnCrossHostInvalidPath()

			Expect(si.GetStatusCode()).To(BeNumerically(">=", 400))
			Expect(si.GetStatusCode()).To(BeNumerically("<", 500))
		})

		Describe("OnCrossHostRef", func() {
			Context("cross-host", func() {
				It("should not trigger error", func() {
					si := NewServeInfo(true, httptest.NewRecorder())
					si.OnCrossHostRef()

					Expect(si.HasError()).To(BeFalse())
				})
			})

			Context("non cross-host", func() {
				It("should trigger error", func() {
					si := NewServeInfo(false, httptest.NewRecorder())
					si.OnCrossHostRef()

					Expect(si.HasError()).To(BeTrue())
				})
			})
		})
	})

	Describe("ResponseWriter", func() {
		It("should add header", func() {
			headerKey := "Key"
			headerValue := "Value"

			si, w := newServeInfo()
			si.AddHeader(headerKey, headerValue)
			si.Flush()

			Expect(w.Header().Get(headerKey)).To(Equal(headerValue))
		})

		It("should set header content length", func() {
			cl := int64(1)

			si, w := newServeInfo()
			si.SetContentLength(cl)
			si.Flush()

			Expect(w.Header().Get("Content-Length")).To(Equal(fmt.Sprintf("%d", cl)))
		})

		It("should write body", func() {
			bytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}

			si, w := newServeInfo()
			si.WriteBody(bytes)

			Expect(w.Header().Get("Content-Length")).To(Equal(fmt.Sprintf("%d", len(bytes))))
			Expect(w.Body.Bytes()).To(Equal(bytes))
		})

		It("should copy body", func() {
			slice := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
			buffer := bytes.NewBuffer(slice)

			si, w := newServeInfo()
			si.SetContentLength(int64(len(slice)))
			si.CopyBody(buffer)

			Expect(w.Header().Get("Content-Length")).To(Equal(fmt.Sprintf("%d", len(slice))))
			Expect(w.Body.Bytes()).To(Equal(slice))
		})

		It("should copy body (length=0)", func() {
			var slice []byte
			buffer := bytes.NewBuffer(slice)

			si, w := newServeInfo()
			si.CopyBody(buffer)

			Expect(w.Header().Get("Content-Length")).To(Equal(""))
			Expect(w.Body.Bytes()).To(BeNil())
		})

		It("should copy body (EOF)", func() {
			var slice []byte
			buffer := bytes.NewBuffer(slice)

			si, _ := newServeInfo()
			si.SetContentLength(int64(1))
			si.CopyBody(buffer)

			t, e := si.GetError()
			Expect(t).To(Equal(int(ErrorCopyBody)))
			Expect(e).To(HaveOccurred())
		})
	})

})
