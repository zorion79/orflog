# orflog - get logs from [ORF Fusion](https://vamsoft.com/)

[![Build Status](https://travis-ci.com/zorion79/orflog.svg?branch=master)](https://travis-ci.com/zorion79/orflog)
[![Coverage Status](https://coveralls.io/repos/github/zorion79/orflog/badge.svg?branch=master)](https://coveralls.io/github/zorion79/orflog?branch=master)

This library get records from orf logs.

- Collect records from many servers in one request.
- Return two channels with new and old records

First you get slice with one month logs.  
Next you will get channel with new logs.

## Install

`go get -u github.com/zorion79/orflog`

## Usage

- define options `Opts`
- memory!!!
- make service `NewService(opts Opts)`
- get channels `s.Channel() (new <-chan *Orf, remove <-chan *Orf)`
