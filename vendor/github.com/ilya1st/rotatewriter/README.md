# Rotation log writer

[![Build Status](https://travis-ci.org/ilya1st/rotatewriter.svg?branch=master)](https://travis-ci.org/ilya1st/rotatewriter)[![Go Report Card](https://goreportcard.com/badge/github.com/ilya1st/rotatewriter)](https://goreportcard.com/report/github.com/ilya1st/rotatewriter)

This is log writer to support rotation.

Note: logs directory in library intended for testing purposes

## Usage

### Example

You can use RotateWriter as standard golang log writer or use as zerolog writer.

Example for zerolog case:

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/ilya1st/rotatewriter"
    "github.com/rs/zerolog"
)

func main() {
    // this is for test
    writer, err := rotatewriter.NewRotateWriter("./logs/test.log", 8)
    if err != nil {
        panic(err)
    }
    sighupChan := make(chan os.Signal, 1)
    signal.Notify(sighupChan, syscall.SIGHUP)
    go func() {
        for {
            _, ok := <-sighupChan
            if !ok {
                return
            }
            fmt.Println("Log rotation")
            writer.Rotate(nil)
        }
    }()
    logger := zerolog.New(writer).With().Timestamp().Logger()
    fmt.Println("Just run in another console and look into logs directory:\n$ killall -HUP rotateexample")
    for {
        logger.Info().Msg("test")
        time.Sleep(500 * time.Millisecond)
    }
}
```

### rotatewriter.NewRotateWriter() arguments

First argument is full log filename.

Second argument - number of files to store.

If it's 0 - writer just reopen file. This is for cases you use logrotate instrument around your app

### rotatewriter.Rotate()

You call it when you need rotation. It has specific locks do not run rotation twice in same moment.
you may pass callback tu function to know rotation is ready.

### rotatewriter.RotationInProgress()

You can determine is rotation operation in progress or not if you need

## Why

This package was inspired by lumberjack - but there is an issue with blocking while log rotation and it's overloaded with features as size counting, gzip but I need just correctly handle reopen.

## Install

```bash
go get github.com/ilya1st/rotatewriter
```