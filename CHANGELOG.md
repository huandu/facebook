# Change Log

## v2.5.5

- `[FIX]` [#176](https://github.com/huandu/facebook/issue/176) Fix a bug that the `UsageInfo` in `PageResult` is not initialized in the first page. Thanks [@uginroot](https://github.com/uginroot).

Note that the tag `v2.5.4` was tagged on a wrong branch. I deleted it to avoid potential issues.

## v2.5.3

- `[NEW]` [#163](https://github.com/huandu/facebook/issue/163) Add a new type `UnmarshalError` to hold error details when `json.Decoder` fails to parse Facebook API response. Thanks [@stlimtat](https://github.com/stlimtat).

## v2.5.2

- `[FIX]` Update `go.mod` import path from `github.com/huandu/facebook` to `github.com/huandu/facebook/v2`. Since go1.14, incompatible versions are omitted by default.

## v2.5.1

- `[FIX]` [#150](https://github.com/huandu/facebook/pull/150) Add additional error messages to help debug issues. Thanks [@sothychan](https://github.com/sothychan).

## v2.5.0

- `[NEW]` [#147](https://github.com/huandu/facebook/issue/147) `MakeParams` is aware of struct field tag `"facebook"` and `"json"` now. It works quite similar to `json.Marshal` except that it makes any data to `Params` instead of JSON string. Thank [@samber](https://github.com/samber) for your input.
- `[FIX]` [#148](https://github.com/huandu/facebook/issue/148) Fix a panic in `Param#Encode` when a nil value is included in `Param`. Thanks again, [@samber](https://github.com/samber).

## v2.4.0

- `[NEW]` [#126](https://github.com/huandu/facebook/pull/126) [#128](https://github.com/huandu/facebook/pull/128) Support v3.3 rate limiting header `x-business-use-case-usage`, `x-ad-account-usage` and `X-FB-Ads-Insights-Throttle` in `UsageInfo`. Use it in any `Result` by calling `Result#UsageInfo()`. It's also available in `PagingResult` by calling `PagingResult#UsageInfo()`. Thanks [@OwlLaboratory](https://github.com/OwlLaboratory) for raising this change for me. And Thanks
  [@AlphaB](https://github.com/AlphaB) for your PR.

## v2.3.3

- `[FIX]` [#120](https://github.com/huandu/facebook/pull/120) Use timing-safe `hmac.Equal` to check signed request. Thanks, [@gstvg](https://github.com/gstvg).

## v2.3.2

- `[FIX]` [#114](https://github.com/huandu/facebook/pull/114) Correctly parse query string in the path which doesn't start with '/', e.g. `fb.Get("me?fields=name,email", nil)`.

## v2.3.1

- `[FIX]` [#114](https://github.com/huandu/facebook/pull/114) Query string in the path, e.g. `fb.Get("/me?fields=name,email", nil)`, works as expected now. Thanks, [@AsifArko](https://github.com/AsifArko).

## v2.3.0

- `[FIX]` [#110](https://github.com/huandu/facebook/pull/110) Use HTTP GET to send request in which the method is `GET`. Thank [@nayakravi](https://github.com/nayakravi) for raising this issue, and Thank [@AlphaB](https://github.com/AlphaB) and [@robbiet480](https://github.com/robbiet480) for your valuable inputs.

## v2.2.0

- `[NEW]` [#103](https://github.com/huandu/facebook/pull/103) Add `RFC3339Timestamps` flag to set `date_format` on every API request. If this flag is set, all date value returned by facebook will be encoded with a format supported by `json.Unmarshal`. Thanks, [@robbiet480](https://github.com/robbiet480).
- `[NEW]` [#105](https://github.com/huandu/facebook/pull/105) Added ability to override `Session` base URL. It's designed for unit testing. All session requests can be redirected to a test server by setting `Session#BaseURL` to a test URL. Thanks, [@vania-pooh](https://github.com/vania-pooh).

## v2.1.2

- `[FIX]` [#87](https://github.com/huandu/facebook/issues/87) Fix a crash in `Session#addUsageInfo`. Thanks, [@flemeur](https://github.com/flemeur).

## v2.1.1

- `[NEW]` [#86](https://github.com/huandu/facebook/pull/86) Parse `X-App-Usage` and `X-Page-Usage` in response header and store the usage information in `Result`. Use `Result#UsageInfo()` to read it. Thanks, [@robbiet480](https://github.com/robbiet480).

## v2.1.0

- `[NEW]` [#81](https://github.com/huandu/facebook/pull/81) Compatible with the struct field's tag used by `json.Unmarshal`. The "json" key works as expected now. If both the "facebook" key and the "json" key exist, use "facebook".

## v2.0.0

- `[NEW]` [#80](https://github.com/huandu/facebook/pull/80) [#71](https://github.com/huandu/facebook/pull/71) All `Session` API, which sends requests to Facebook, support `Context` now. Thanks, [@sebnow](https://github.com/sebnow) for your thoughts and reminder.
- `[NEW]` [#79](https://github.com/huandu/facebook/pull/79) Add some number types which can be decoded from a string implicitly.
- `[NEW]` [#78](https://github.com/huandu/facebook/pull/78) [#57](https://github.com/huandu/facebook/issues/57) Deprecate FQL and remove all related code.
- `[FIX]` [#73](https://github.com/huandu/facebook/pull/73) Fix regular expression for video post. Thanks, [@acochrane](https://github.com/acochrane).
- `[FIX]` [#62](https://github.com/huandu/facebook/pull/62) Use `base64.RawURLEncoding` to decode signed request data. Thanks,[@zonr](https://github.com/zonr).
- `[FIX]` Fix some typos in README and test cases. Thank [@nick3399](https://github.com/nick3399), [@J-P-77](https://github.com/J-P-77), [@smasher164](https://github.com/smasher164), [@enm10k](https://github.com/enm10k) and many others.
- `[FIX]` Clean up code for readability.

## v1.8.1

- `[FIX]` [#60](https://github.com/huandu/facebook/pull/60) Handle string errors in `Decode()`. Thanks, [@sebnow](https://github.com/sebnow).

## v1.8.0

- `[FIX]` [#59](https://github.com/huandu/facebook/pull/59) Guess content type for binary params by filename extension or an arbitrary value. Thanks, [@panki](https://github.com/panki).

## v1.7.1

- `[FIX]` Fix a tiny bug which slightly affects performance when decoding anonymous field.

## v1.7.0

- `[NEW]` [#50](https://github.com/huandu/facebook/issues/50) `Result` can decode embedded struct field now.
- `[NEW]` Add a new field tag `facebook:"-"` to omit the field when decoding. It can improve decoding performance slightly.

## v1.6.0

- `[NEW]` [#42](https://github.com/huandu/facebook/issues/42) Support custom JSON unmarshaling and `json.Unmarshaler` interface in decoding.

## v1.5.6

- `[NEW]` [#40](https://github.com/huandu/facebook/issues/40) `Session` works with http client created by package `golang.org/x/oauth2`. README is updated with a sample.

## v1.5.5

- `[FIX]` [#39](https://github.com/huandu/facebook/issues/39) When `/oauth/access_token` returns a query string, this package can parse `expires` or `expires_in` field correctly.

## v1.5.4

- `[FIX]` [#37](https://github.com/huandu/facebook/issues/37) Add missing `client_secret` in query string when parsing client code.

## v1.5.3

- `[FIX]` [#34](https://github.com/huandu/facebook/issues/34) Use `expires` instead of `expires_in` if possible when exchanging token or parsing code.

## v1.5.2

- `[FIX]` [#32](https://github.com/huandu/facebook/issues/32) BatchApi/Batch returns facebook error when access token is not valid.

## v1.5.1

- `[FIX]` [#31](https://github.com/huandu/facebook/issues/31) When `/oauth/access_token` returns a query string instead of json, this package can correctly handle it.

## v1.5.0

- `[NEW]` [#28](https://github.com/huandu/facebook/issues/28) Support debug mode introduced by facebook graph API v2.3.
- `[FIX]` Removed all test cases depending on facebook graph API v1.0.

## v1.4.1

- `[NEW]` [#27](https://github.com/huandu/facebook/pull/27) Timestamp value in Graph API response can be decoded as a `time.Time` value now. Thanks, [@Lazyshot](https://github.com/Lazyshot).

## v1.4.0

- `[FIX]` [#23](https://github.com/huandu/facebook/issues/24) Algorithm change: Camel case string to underscore string supports abbreviation

Fix for [#23](https://github.com/huandu/facebook/issues/24) could be a breaking change. Camel case string `HTTPServer` will be converted to `http_server` instead of `h_t_t_p_server`. See issue description for detail.

## v1.3.0

- `[NEW]` [#22](https://github.com/huandu/facebook/issues/22) Add a new helper struct `BatchResult` to hold batch request responses.

## v1.2.0

- `[NEW]` [#20](https://github.com/huandu/facebook/issues/20) Add Decode functionality for paging results. Thanks, [@cbroglie](https://github.com/cbroglie).
- `[FIX]` [#21](https://github.com/huandu/facebook/issues/21) `Session#Inspect` cannot return error if access token is invalid.

Fix for [#21](https://github.com/huandu/facebook/issues/21) will result a possible breaking change in `Session#Inspect`. It was return whole result returned by facebook inspect api. Now it only return its "data" sub-tree. As facebook puts everything including error message in "data" sub-tree, I believe it's reasonable to make this change.

## v1.1.0

- `[FIX]` [#19](https://github.com/huandu/facebook/issues/19) Any valid int64 number larger than 2^53 or smaller than -2^53 can be correctly decoded without precision lost.

Fix for [#19](https://github.com/huandu/facebook/issues/19) will result a possible breaking change in `Result#Get` and `Result#GetField`. If a JSON field is a number, these two functions will return `json.Number` instead of `float64`.

The fix also introduces a side effect in `Result#Decode` and `Result#DecodeField`. A number field (`int*` and `float*`) can be decoded to a string. It was not allowed in previous version.

## v1.0.0

Initial tag. Library is stable enough for all features mentioned in README.md.
