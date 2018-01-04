build:
	gox -output="release/check_consul_service_{{.OS}}_{{.Arch}}" -os="darwin linux" -arch="amd64" -ldflags "-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=`git rev-parse HEAD` -X main.version=`cat VERSION`"
