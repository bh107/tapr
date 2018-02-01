# Tapr dev notes

## Running with LTFS emulation
Type `vagrant up` to create a CentOS 7 VirtualBox VM with LTFS installed.

Install the vbguest Vagrant plugin for a better experience:

    vagrant plugin install vagrant-vbguest

## LTFS
Get the LTFS SDE from IBM and place it into the `./blobs` directory.

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
