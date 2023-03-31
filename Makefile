
run: 
	go run main.go \
		--channelid UCp-5t9SrOQwXMU7iIjQfARg \
		--playlistid PLAGlaP7ijaoVpQ9W2I9ixQ77QjICt-NRw \
		--webhook "https://discord.com/api/webhooks/1090629278997151764/8Y3FRQIemLcIVPn3bYGeaLFWP9xvMPxhEIaVdU797FgoLiy6W1fteVCR-w5fn2vlrVTy"

init:
	go run main.go \
		--phase init \
		--webhook "https://discord.com/api/webhooks/1090629278997151764/8Y3FRQIemLcIVPn3bYGeaLFWP9xvMPxhEIaVdU797FgoLiy6W1fteVCR-w5fn2vlrVTy"

build:
	go build -o bin/ main.go

image-build:
	docker build -t github.com/ryusa/asamio-upodater .
