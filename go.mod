module http-mqtt-boilerplate

go 1.25.6

require (
	github.com/amacneil/dbmate/v2 v2.29.3
	github.com/coder/guts v1.6.3
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/getkin/kin-openapi v0.133.0
	github.com/go-chi/chi/v5 v5.2.4
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.33
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037
	golang.org/x/text v0.33.0
	golang.org/x/tools v0.41.0
)

require (
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/dop251/goja v0.0.0-20241024094426-79f3a7efcdbd // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/google/pprof v0.0.0-20230207041349-798e818bf904 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/woodsbury/decimal128 v1.4.0 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)

// Using fork with OpenAPI 3.1 support (PR #1114: https://github.com/getkin/kin-openapi/pull/1114)
// Upstream issue #230 has been open since 2020. This fork adds:
// - OpenAPI 3.1.0 support (type arrays, const, JSON Schema 2020-12)
// Can be removed once upstream merges PR #1114
replace github.com/getkin/kin-openapi => github.com/RobbertDM/kin-openapi-fork v0.0.0-20260113132111-14628ad5f844
