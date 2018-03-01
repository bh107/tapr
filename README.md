# Tapr

Documentation: [tapr.space](https://tapr.space)

## About

Tapr is an experimental high performance tape management system.

## Status

Tapr is not ready for production use.

## Usage

### Installation

Tapr can be installed using the standard Go tool chain.

    go get tapr.space/cmd/...


### Server

    taprd -log debug -audit -dbreset


### Client

    tapr put <big-file>

    tapr get -dest /tmp <big-file>


