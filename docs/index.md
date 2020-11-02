<img src="assets/logo.png" alt="FTPGrab" width="128px" style="display: block; margin-left: auto; margin-right: auto"/>

<p align="center">
  <a href="https://github.com/crazy-max/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/ftpgrab/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/ftpgrab/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/ftpgrab/actions?workflow=build"><img src="https://img.shields.io/github/workflow/status/crazy-max/ftpgrab/build?label=build&logo=github&style=flat-square" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/ftpgrab/"><img src="https://img.shields.io/docker/stars/crazymax/ftpgrab.svg?style=flat-square&logo=docker" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/ftpgrab/"><img src="https://img.shields.io/docker/pulls/crazymax/ftpgrab.svg?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/ftpgrab"><img src="https://goreportcard.com/badge/github.com/crazy-max/ftpgrab?style=flat-square" alt="Go Report"></a>
  <a href="https://app.codacy.com/gh/crazy-max/ftpgrab"><img src="https://img.shields.io/codacy/grade/5d94f58df1b34c238e26db6a52cb92a0.svg?style=flat-square" alt="Code Quality"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

---

## What is FTPGrab?

**FTPGrab** :zap: is a CLI application written in [Go](https://golang.org/) and delivered as a
[single executable]({{ config.repo_url }}releases/latest) (and a [Docker image](install/docker.md))
to grab your files from a remote FTP or SFTP server to your NAS, server or computer.

With Go, this can be done with an independent binary distribution across all platforms and architectures that Go supports.
This support includes Linux, macOS, and Windows, on architectures like amd64, i386, ARM, PowerPC, and others.

Because FTPGrab is distributed as an independent binary, it is ideal for those with a seedbox to grab your files
periodically to your Synology, Qnap, D-Link and others NAS.

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

## License

This project is licensed under the terms of the MIT license.
