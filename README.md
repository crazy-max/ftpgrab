<p align="center"><a href="https://ftpgrab.github.io" target="_blank"><img width="100" src="https://ftpgrab.github.io/img/logo.png"></a></p>

<p align="center">
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/ftpgrab/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/ftpgrab/ftpgrab/releases/latest"><img src="https://img.shields.io/github/downloads/ftpgrab/ftpgrab/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/ftpgrab/ftpgrab/actions"><img src="https://github.com/ftpgrab/ftpgrab/workflows/build/badge.svg" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/ftpgrab/ftpgrab/"><img src="https://img.shields.io/docker/stars/ftpgrab/ftpgrab.svg?style=flat-square" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/ftpgrab/ftpgrab/"><img src="https://img.shields.io/docker/pulls/ftpgrab/ftpgrab.svg?style=flat-square" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/ftpgrab/ftpgrab"><img src="https://goreportcard.com/badge/github.com/ftpgrab/ftpgrab?style=flat-square" alt="Go Report"></a>
  <a href="https://www.codacy.com/app/ftpgrab/ftpgrab"><img src="https://img.shields.io/codacy/grade/354bfb181fc5482dac1e8f31e8e29af5.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

## About

**FTPGrab** :zap: is a CLI application written in [Go](https://golang.org/) to grab your files from a remote FTP or SFTP server to your NAS, server or computer. With Go, this app can be used across many platforms and architectures. This support includes Linux, FreeBSD, macOS and Windows on architectures like amd64, i386, ARM and others.

Because FTPGrab is distributed as an independent binary, it is ideal for those with a seedbox to grab your files periodically to your Synology, Qnap, D-Link and others NAS.

![](https://ftpgrab.github.io/img/faq/ftpgrab-report-slack.png)

## Features

* Multiple sources
* SFTP support
* Prevent re-download through a hash
* Efficient key/value store database to audit files already downloaded
* Internal cron implementation through go routines
* Include and exclude filters with regular expression
* Date filter
* Retry on failed download
* Change file/folder permissions and owner
* Translate modtimes on downloaded files
* Beautiful email report
* Webhook notification
* Slack incoming webhook notification
* Enhanced logging
* Timezone can be changed
* Official Docker image

## Documentation

* [Get started](https://ftpgrab.github.io/doc/get-started/)
* Installation
  * [With Docker](https://ftpgrab.github.io/doc/install-with-docker/)
  * [From binary](https://ftpgrab.github.io/doc/install-from-binary/)
  * [Linux service](https://ftpgrab.github.io/doc/linux-service/)
* [Configuration](https://ftpgrab.github.io/doc/configuration/)
* [FAQ](https://ftpgrab.github.io/doc/faq/)
* [Changelog](https://ftpgrab.github.io/doc/changelog/)
* [Upgrade notes](https://ftpgrab.github.io/doc/upgrade-notes/)
* [Reporting an issue](https://ftpgrab.github.io/doc/reporting-issue/)

## How can I help ?

All kinds of contributions are welcome :raised_hands:! The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon: You can also support this project by [**becoming a sponsor on GitHub**](https://github.com/sponsors/crazy-max) :clap: or by making a [Paypal donation](https://www.paypal.me/crazyws) to ensure this journey continues indefinitely! :rocket:

Thanks again for your support, it is much appreciated! :pray:

## License

MIT. See `LICENSE` for more details.<br />
Icon credit to [Nick Roach](http://www.elegantthemes.com/).
