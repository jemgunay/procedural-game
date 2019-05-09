APP_NAME="procedural-game"

# Clean up generated files.
clean:
	rm -rf ${APP_NAME} ./build

# Build game executable.
build: clean
	go build

# Build game into zip archive.
package: build
	mkdir ./build
	zip -r ${APP_NAME}.zip ${APP_NAME} ./assets
	mv ${APP_NAME}.zip ./build
