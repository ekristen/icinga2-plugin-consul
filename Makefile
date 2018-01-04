build:
	gox -output="release/{{.Dir}}_{{.OS}}_{{.Arch}}" -ldflags "-X main.buildstamp `date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash `git rev-parse HEAD`"
