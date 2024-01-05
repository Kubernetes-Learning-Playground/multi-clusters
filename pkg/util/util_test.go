package util

import (
	"fmt"
	"testing"
)

func TestParseIntoGvr(t *testing.T) {
	fmt.Println(ParseIntoGvr("api.practice.com/v1alpha1/proxys", "/"))
}
