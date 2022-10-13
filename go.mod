module github.com/ONSdigital/dp-api-router

go 1.18

replace (
	// solves CVE-2021-3121
	github.com/prometheus/common => github.com/prometheus/common v0.34.0
	// solves sonatype-2020-0584 CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting') github.com/yuin/goldmark - Cross-Site Scripting (XSS)
	github.com/yuin/goldmark => github.com/yuin/goldmark v1.4.12
)

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.162.0
	github.com/ONSdigital/dp-authorisation/v2 v2.25.1
	github.com/ONSdigital/dp-healthcheck v1.3.0
	github.com/ONSdigital/dp-kafka/v2 v2.5.0
	github.com/ONSdigital/dp-kafka/v3 v3.3.2
	github.com/ONSdigital/dp-net/v2 v2.4.0
	github.com/ONSdigital/go-ns v0.0.0-20210916104633-ac1c1c52327e
	github.com/ONSdigital/log.go/v2 v2.2.0
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/justinas/alice v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.7.2
)

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0 // indirect
	github.com/ONSdigital/dp-net v1.5.0 // indirect
	github.com/ONSdigital/dp-rchttp v1.0.0 // indirect
	github.com/Shopify/sarama v1.36.0 // indirect
	github.com/aws/aws-sdk-go v1.44.76 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-avro/avro v0.0.0-20171219232920-444163702c11 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/net v0.0.0-20220812174116-3211cb980234 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/text v0.3.8 // indirect
)
