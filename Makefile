build:
	cd ./go; go build -o ../bin/NvimSpotify ./plugin

manifest:
	NvimSpotify -manifest lspmeta

clean:
	rm -rf bin/NvimSpotify
