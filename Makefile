build:
	cd ./go; go build -o ../bin/NvimSpotify

manifest:
	NvimSpotify -manifest NvimSpotify

clean:
	rm -rf bin/NvimSpotify
