module github.com/uwu-tools/peribolos

go 1.18

// Upstream is unmaintained. This fork introduces two important changes:
// - We log an error if writing a cache key fails e.g., because disk is full
// - We inject a header that allows ghproxy to detect if the response was revalidated or a cache miss
replace github.com/gregjones/httpcache => github.com/alvaroaleman/httpcache v0.0.0-20210618195546-ab9a1a3f8a38

require (
	github.com/airconduct/go-probot v0.0.3
	github.com/caarlos0/env/v7 v7.1.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/ghodss/yaml v1.0.0
	github.com/gomodule/redigo v1.8.9
	github.com/google/go-cmp v0.5.9
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/prometheus/client_golang v1.16.0
	github.com/sethvargo/go-githubactions v1.1.0
	github.com/shurcooL/githubv4 v0.0.0-20230305132112-efb623903184
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/oauth2 v0.10.0
	golang.org/x/sync v0.3.0
	k8s.io/apimachinery v0.27.3
	k8s.io/utils v0.0.0-20230308161112-d77c459e9343
	sigs.k8s.io/release-utils v0.7.4
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradleyfalzon/ghinstallation/v2 v2.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be // indirect
	github.com/emicklei/go-restful-openapi/v2 v2.9.1 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/spec v0.20.8 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-github/v45 v45.2.0 // indirect
	github.com/google/go-github/v48 v48.2.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.1-0.20210504230335-f78f29fc09ea // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sethvargo/go-envconfig v0.8.0 // indirect
	github.com/shurcooL/graphql v0.0.0-20220606043923-3cf50f8a0a29 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xanzy/go-gitlab v0.78.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
