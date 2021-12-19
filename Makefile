build:
	cp ./go/nvim-telescope /usr/local/bin

gobuild:
	cd ./go; go build -o ./nvim-telescope ./plugin

manifest:
	nvim-spotify -manifest lspmeta

clean:
	rm ./go/nvim-telescope
	rm -rf /usr/local/bin/nvim-spotify
