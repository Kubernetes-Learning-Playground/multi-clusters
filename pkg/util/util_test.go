package util

import (
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestParseIntoGvr(t *testing.T) {
	Convey("Parse to GVR test", t, func() {
		resTest := schema.GroupVersionResource{
			Group:    "api.practice.com",
			Version:  "v1alpha1",
			Resource: "tests",
		}
		res, _ := ParseIntoGvr("api.practice.com/v1alpha1/tests", "/")
		So(res, ShouldEqual, resTest)

		pods := schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		}
		res, _ = ParseIntoGvr("core/v1/pods", "/")
		So(res, ShouldEqual, pods)

		_, err := ParseIntoGvr("pods", "/")
		So(err, ShouldNotBeNil)

	})
}
