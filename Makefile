include config.env

run: 
	go run main.go \
		--channelid UCp-5t9SrOQwXMU7iIjQfARg \
		--playlistid PLAGlaP7ijaoVpQ9W2I9ixQ77QjICt-NRw \
		--webhook "$(DISCORD_WEBHOOK)"

init:
	go run main.go \
		--phase init \
		--webhook "$(DISCORD_WEBHOOK)"

build:
	go build -o bin/main main.go

image-build:
	docker build -t ghcr.io/ryusa/asamio-updater .
