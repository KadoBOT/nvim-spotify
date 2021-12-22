build:
	cd ./go; go build -o ../bin/NvimSpotify ./plugin

manifest:
	NvimSpotify -manifest NvimSpotify

clean:
	rm -rf bin/NvimSpotify
