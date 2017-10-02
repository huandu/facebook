// A facebook graph api client in go.
// https://github.com/huandu/facebook/
//
// Copyright 2012 - 2015, Huan Du
// Licensed under the MIT license
// https://github.com/huandu/facebook/blob/master/LICENSE

package facebook

import (
	"testing"
)

func TestCamelCaseToUnderScore(t *testing.T) {
	cases := map[string]string{
		"TestCase":           "test_case",
		"HTTPServer":         "http_server",
		"NoHTTPS":            "no_https",
		"Wi_thF":             "wi_th_f",
		"_AnotherTES_TCaseP": "_another_tes_t_case_p",
		"ALL":                "all",
		"UserID":             "user_id",
	}

	for k, v := range cases {
		str := camelCaseToUnderScore(k)

		if str != v {
			t.Fatalf("wrong underscore string. [expect:%v] [actual:%v]", v, str)
		}
	}
}
