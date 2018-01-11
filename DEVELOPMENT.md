# Tapr dev notes

## Update dependencies
To update the dep-managed dependencies for the project do

    $ dep ensure -update
    $ dep prune
    $ find vendor -name '*_test.go' -delete


# Generate
To generate protocol buffers etc do

    $ go generate ./...


# Run tests

    $ go test ./...
