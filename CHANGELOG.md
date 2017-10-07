# Change Log #

## v2.1.0 ##

* `[NEW]` [#81](https://github.com/huandu/facebook/pull/81) Compatible with the struct field's tag used by `json.Unmarshal`. The "json" key works as expected now. If both the "facebook" key and the "json" key exist, use "facebook".

## v2.0.0 ##

* `[NEW]` [#80](https://github.com/huandu/facebook/pull/80) [#71](https://github.com/huandu/facebook/pull/71) All `Session` API, which sends requests to Facebook, support `Context` now. Thanks [@sebnow](https://github.com/sebnow) for your thoughts and reminder.
* `[NEW]` [#79](https://github.com/huandu/facebook/pull/79) Add some number types which can be decoded from a string implicitly.
* `[NEW]` [#78](https://github.com/huandu/facebook/pull/78) [#57](https://github.com/huandu/facebook/issues/57) Deprecate FQL and remove all related code.
* `[FIX]` [#73](https://github.com/huandu/facebook/pull/73) Fix regular expression for video post. Thanks [@acochrane](https://github.com/acochrane).
* `[FIX]` [#62](https://github.com/huandu/facebook/pull/62) Use base64.RawURLEncoding to decode signed request data. Thanks [@zonr](https://github.com/zonr).
* `[FIX]` Fix some typos in README and test cases. Thanks [@nick3399](https://github.com/nick3399), [@J-P-77](https://github.com/J-P-77), [@smasher164](https://github.com/smasher164), [@enm10k](https://github.com/enm10k) and many others. Thank you.
* `[FIX]` Clean up code for readability. 

## v1.8.1 ##

* `[FIX]` [#60](https://github.com/huandu/facebook/pull/60) Handle string errors in `Decode()`. Thanks [@sebnow](https://github.com/sebnow).

## v1.8.0 ##

* `[FIX]` [#59](https://github.com/huandu/facebook/pull/59) Guess content type for binary params by filename extension or an arbitrary value. Thanks [@panki](https://github.com/panki).

## v1.7.1 ##

* `[FIX]` Fix a tiny bug which slightly affects performance when decoding anonymous field.

## v1.7.0 ##

* `[NEW]` [#50](https://github.com/huandu/facebook/issues/50) `Result` can decode embedded struct field now.
* `[NEW]` Add a new field tag `facebook:"-"` to omit the field when decoding. It can improve decoding performance slightly.

## v1.6.0 ##

* `[NEW]` [#42](https://github.com/huandu/facebook/issues/42) Support custom JSON unmarshaling and `json.Unmarshaler` interface in decoding.

## v1.5.6 ##

* `[NEW]` [#40](https://github.com/huandu/facebook/issues/40) `Session` works with http client created by package `golang.org/x/oauth2`. README is updated with a sample.

## v1.5.5 ##

* `[FIX]` [#39](https://github.com/huandu/facebook/issues/39) When `/oauth/access_token` returns a query string, this package can parse `expires` or `expires_in` field correctly.

## v1.5.4 ##

* `[FIX]` [#37](https://github.com/huandu/facebook/issues/37) Add missing `client_secret` in query string when parsing client code.

## v1.5.3 ##

* `[FIX]` [#34](https://github.com/huandu/facebook/issues/34) Use `expires` instead of `expires_in` if possible when exchanging token or parsing code.

## v1.5.2 ##

* `[FIX]` [#32](https://github.com/huandu/facebook/issues/32) BatchApi/Batch returns facebook error when access token is not valid.

## v1.5.1 ##

* `[FIX]` [#31](https://github.com/huandu/facebook/issues/31) When `/oauth/access_token` returns a query string instead of json, this package can correctly handle it.

## v1.5.0 ##

* `[NEW]` [#28](https://github.com/huandu/facebook/issues/28) Support debug mode introduced by facebook graph API v2.3.
* `[FIX]` Removed all test cases depending on facebook graph API v1.0.

## v1.4.1 ##

* `[NEW]` [#27](https://github.com/huandu/facebook/pull/27) Timestamp value in Graph API response can be decoded as a `time.Time` value now. Thanks, [@Lazyshot](https://github.com/Lazyshot).

## v1.4.0 ##

* `[FIX]` [#23](https://github.com/huandu/facebook/issues/24) Algorithm change: Camel case string to underscore string supports abbreviation

Fix for [#23](https://github.com/huandu/facebook/issues/24) could be a breaking change. Camel case string `HTTPServer` will be converted to `http_server` instead of `h_t_t_p_server`. See issue description for detail.

## v1.3.0 ##

* `[NEW]` [#22](https://github.com/huandu/facebook/issues/22) Add a new helper struct `BatchResult` to hold batch request responses.

## v1.2.0 ##

* `[NEW]` [#20](https://github.com/huandu/facebook/issues/20) Add Decode functionality for paging results. Thanks, [@cbroglie](https://github.com/cbroglie).
* `[FIX]` [#21](https://github.com/huandu/facebook/issues/21) `Session#Inspect` cannot return error if access token is invalid.

Fix for [#21](https://github.com/huandu/facebook/issues/21) will result a possible breaking change in `Session#Inspect`. It was return whole result returned by facebook inspect api. Now it only return its "data" sub-tree. As facebook puts everything including error message in "data" sub-tree, I believe it's reasonable to make this change.

## v1.1.0 ##

* `[FIX]` [#19](https://github.com/huandu/facebook/issues/19) Any valid int64 number larger than 2^53 or smaller than -2^53 can be correctly decoded without precision lost.

Fix for [#19](https://github.com/huandu/facebook/issues/19) will result a possible breaking change in `Result#Get` and `Result#GetField`. If a JSON field is a number, these two functions will return `json.Number` instead of `float64`.

The fix also introduces a side effect in `Result#Decode` and `Result#DecodeField`. A number field (`int*` and `float*`) can be decoded to a string. It was not allowed in previous version.

## v1.0.0 ##

Initial tag. Library is stable enough for all features mentioned in README.md.
