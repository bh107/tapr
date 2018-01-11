# Tapr

Documentation: [tapr.space](https://hpt.space/tapr/)

## About

Tapr is an experimental high performance tape management system.

## Status

Tapr is not ready for production use.

## Usage

### Server

   taprd -log debug -audit -dbreset


### Client

    tapr put <big-file>

    tapr get -dest /tmp <big-file>


### Administrative client

Tail logs
    tapr log -tail
