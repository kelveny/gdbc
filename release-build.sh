go build -v -ldflags="-X main.SemVer=`git describe --tags --abbrev=0`"
