build:
	cp ./go/nvim-spotify /usr/local/bin

gobuild:
	cd ./go; go build -o ./nvim-spotify ./plugin

manifest:
	nvim-spotify -manifest lspmeta

clean:
	rm ./go/nvim-telescope
	rm -rf /usr/local/bin/nvim-spotify
