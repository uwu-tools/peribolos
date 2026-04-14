module github.com/uwu-tools/peribolos

go 1.25.0

// Upstream is unmaintained. This fork introduces two important changes:
// - We log an error if writing a cache key fails e.g., because disk is full
// - We inject a header that allows ghproxy to detect if the response was revalidated or a cache miss
replace github.com/gregjones/httpcache => github.com/alvaroaleman/httpcache v0.0.0-20210618195546-ab9a1a3f8a38

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/caarlos0/env/v7 v7.1.0
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gomodule/redigo v1.9.3
	github.com/google/go-cmp v0.7.0
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/prometheus/client_golang v1.23.2
	github.com/sethvargo/go-githubactions v1.3.2
	github.com/shurcooL/githubv4 v0.0.0-20230305132112-efb623903184
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	go4.org v0.0.0-20230225012048-214862532bf5
	golang.org/x/oauth2 v0.36.0
	golang.org/x/sync v0.20.0
	k8s.io/apimachinery v0.35.3
	k8s.io/utils v0.0.0-20251002143259-bc988d571ff4
	sigs.k8s.io/release-utils v0.12.4
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/shurcooL/graphql v0.0.0-20220606043923-3cf50f8a0a29 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250910181357-589584f1c912 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)
