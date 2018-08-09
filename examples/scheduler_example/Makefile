build:
	cd ../../vendor/github.com/ligato/cn-infra/kvscheduler/descriptor/adapter-generator && go install -v
	cd ifplugin/descriptor && rm -rf adapter/ && go generate
	go build
