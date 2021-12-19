build:
	cp ./go/nvim-telescope /usr/local/bin

manifest:
	nvim-spotify -manifest lspmeta

clean:
	rm -rf /usr/local/bin/nvim-spotify
