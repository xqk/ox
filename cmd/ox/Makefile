version := $(shell /bin/date "+%Y-%m-%d %H:%M")

mac:
	GOOS=darwin go build -ldflags="-s -w" -ldflags="-X 'main.BuildTime=$(version)'" -o ox main.go
	$(if $(shell command -v upx), upx ox)
	mv ox /usr/local/bin/