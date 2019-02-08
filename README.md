<p align="center"><a href="https://ftpgrab.github.io" target="_blank"><img width="100" src="https://ftpgrab.github.io/img/logo.png"></a></p>

<p align="center">
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/ftpgrab/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://travis-ci.com/ftpgrab/ftpgrab"><img src="https://img.shields.io/travis/com/ftpgrab/ftpgrab/master.svg?style=flat-square" alt="Build Status"></a>
  <a href="https://www.codacy.com/app/ftpgrab/ftpgrab"><img src="https://img.shields.io/codacy/grade/354bfb181fc5482dac1e8f31e8e29af5.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL"><img src="https://img.shields.io/badge/donate-paypal-7057ff.svg?style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**FTPGrab** is a CLI application written in [Go](https://golang.org/) distributed through a single binary to grab your files from a remote FTP server to your NAS, server or computer.

This application can be used on many platforms like Linux, MacOS, Windows, Synology, Qnap, D-Link...

Before reporting an issue, please read the [Troubleshooting page](https://ftpgrab.github.io/doc/troubleshooting).

## Features

* Multiple sources
* Prevent re-download through a hash
* Efficient key/value store database to store hash
* Internal cron implementation through go routines
* Include and exclude filters with regular expression
* Date filter
* Retry on failed download
* Change file/folder permissions and owner
* Beautiful email report
* Enhanced logging
* Timezone can be changed
* 🐳 Official [Docker image available](https://hub.docker.com/r/ftpgrab/ftpgrab/). Check [this repo](https://github.com/ftpgrab/docker) for more info

## Documentation

* [Get started](https://ftpgrab.github.io/doc/get-started)
* [Configuration](https://ftpgrab.github.io/doc/configuration)
* [Troubleshooting](https://ftpgrab.github.io/doc/troubleshooting)
* [Changelog](https://ftpgrab.github.io/doc/changelog)
* [Upgrade notes](https://ftpgrab.github.io/doc/upgrade-notes)
* [Reporting an issue](https://ftpgrab.github.io/doc/reporting-issue)

## Logs

```console
Tue, 29 Jan 2019 22:23:58 CET INF Starting FTPGrab 5.0.0
Tue, 29 Jan 2019 22:23:58 CET INF ########
Tue, 29 Jan 2019 22:23:58 CET INF Connecting to 198.51.100.0:21...
Tue, 29 Jan 2019 22:23:58 CET INF Grabbing from /complete/
Tue, 29 Jan 2019 22:23:59 CET INF --------
Tue, 29 Jan 2019 22:23:59 CET INF Checking /complete/Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi
Tue, 29 Jan 2019 22:23:59 CET INF Never downloaded
Tue, 29 Jan 2019 22:23:59 CET INF Downloading file (184.18MB) to /tmp/seedbox/Burn.Notice.S06E16.VOSTFR.HDTV.XviD.avi...
Tue, 29 Jan 2019 22:24:47 CET INF File successfully downloaded!
Tue, 29 Jan 2019 22:24:47 CET INF Time spent: 48 seconds
Tue, 29 Jan 2019 22:24:47 CET INF --------
Tue, 29 Jan 2019 22:24:47 CET INF Checking /complete/Burn.Notice.S06E17.VOSTFR.HDTV.XviD.avi
Tue, 29 Jan 2019 22:24:47 CET INF Never downloaded
Tue, 29 Jan 2019 22:24:47 CET INF Downloading file (186.27MB) to /tmp/seedbox/Burn.Notice.S06E17.VOSTFR.HDTV.XviD.avi...
Tue, 29 Jan 2019 22:25:40 CET INF File successfully downloaded!
Tue, 29 Jan 2019 22:25:40 CET INF Time spent: 50 seconds
Tue, 29 Jan 2019 22:25:40 CET INF ########
Tue, 29 Jan 2019 22:25:41 CET INF Finished, total time spent: 1 minute 49 seconds
```

## How can I help ?

All kinds of contributions are welcome :raised_hands:!<br />
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:<br />
But we're not gonna lie to each other, I'd rather you buy me a beer or two :beers:!

[![Paypal](https://ftpgrab.github.io/img/paypal-donate.png)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL)

## License

MIT. See `LICENSE` for more details.<br />
Icon credit to [Nick Roach](http://www.elegantthemes.com/).
