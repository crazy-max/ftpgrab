<p align="center"><a href="https://ftpgrab.github.io" target="_blank"><img width="100" src="https://ftpgrab.github.io/img/logo.png"></a></p>

<p align="center">
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/ftpgrab/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/downloads/ftpgrab/ftpgrab/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://travis-ci.com/ftpgrab/ftpgrab"><img src="https://img.shields.io/travis/com/ftpgrab/ftpgrab/5.0.x.svg?style=flat-square" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/ftpgrab/ftpgrab"><img src="https://goreportcard.com/badge/github.com/ftpgrab/ftpgrab?style=flat-square" alt="Go Report"></a>
  <a href="https://www.codacy.com/app/ftpgrab/ftpgrab"><img src="https://img.shields.io/codacy/grade/354bfb181fc5482dac1e8f31e8e29af5/5.0.x.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL"><img src="https://img.shields.io/badge/donate-paypal-7057ff.svg?style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**FTPGrab** is a CLI application written in [Go](https://golang.org/) distributed through a single binary to grab ⏬ your files from a remote FTP server to your NAS, server or computer. This application can be used on many platforms like Linux, MacOS, Windows, Synology, Qnap, D-Link...

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
* Translate modtimes on downloaded files
* Beautiful email report
* Enhanced logging
* Timezone can be changed
* 🐳 Official [Docker image available](https://github.com/ftpgrab/docker)

## Documentation

* [Get started](https://ftpgrab.github.io/doc/get-started/)
* [Configuration](https://ftpgrab.github.io/doc/configuration/)
* [FAQ](https://ftpgrab.github.io/doc/faq/)
* [Changelog](https://ftpgrab.github.io/doc/changelog/)
* [Upgrade notes](https://ftpgrab.github.io/doc/upgrade-notes/)
* [Reporting an issue](https://ftpgrab.github.io/doc/reporting-issue/)

## TODO

* [ ] Linux service sample
* [ ] Windows service sample
* [ ] [Chocolatey](https://chocolatey.org/) package
* [ ] [Brew](https://brew.sh/) recipe
* [ ] [Cloudron](https://cloudron.io/) app
* [ ] Sublogger / dictionary for entries
* [ ] Build / Install from source doc

## How can I help ?

All kinds of contributions are welcome :raised_hands:!<br />
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:<br />
But we're not gonna lie to each other, I'd rather you buy me a beer or two :beers:!

[![Paypal](https://ftpgrab.github.io/img/paypal-donate.png)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=7NFD44VBNE3VL)

## License

MIT. See `LICENSE` for more details.<br />
Icon credit to [Nick Roach](http://www.elegantthemes.com/).
