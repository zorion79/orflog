# orflog - get logs from [ORF Fusion](https://vamsoft.com/)

This library get records from orf logs.

- Collect records from many servers in one request.
- Can return slice or channel

## Install

`go get -u github.com/zorion79/orflog`

## Usage

- define options `Opts`
- make service `NewService(opts Opts)`
- get slice `GetLogs() []Orf` or channel `GetLogsChan() <- chan Orf`