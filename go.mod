module gatanity/assam

go 1.19

require (
	github.com/aws/aws-sdk-go v1.44.134
	github.com/chromedp/cdproto v0.0.0-20221108233440-fad8339618ab
	github.com/chromedp/chromedp v0.8.6
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.6.1
	github.com/stretchr/testify v1.8.1
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude (
	// Exclude x/text affected by CVE-2020-28852.
	// https://github.com/golang/go/issues/42536
	golang.org/x/text v0.3.0
	golang.org/x/text v0.3.1
	golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2
	golang.org/x/text v0.3.2
	golang.org/x/text v0.3.3
	golang.org/x/text v0.3.4
)
