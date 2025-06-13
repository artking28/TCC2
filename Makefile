
build:
	cd src && \
	go build -o ../out/server ./cmd/server-go && \
	go build -o ../out/bot ./cmd/scrapper-bot && clear;

server:
	clear && go run src/cmd/scrapper-bot/main.go;

bot:
	clear && go run src/cmd/scrapper-bot/main.go;

clear: 
	rm -r out && clear;