gox -ldflags="-X main.version=$(git describe --always --dirty)" -output="$GOPATH/out/{{.Dir}}/$(git describe --always --dirty)/{{.Dir}}_{{.OS}}_{{.Arch}}"
