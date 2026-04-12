<img src="assets/logo.png" alt="FTPGrab" width="128px" style="display: block; margin-left: auto; margin-right: auto"/>

<p align="center">
  <a href="https://github.com/crazy-max/ftpgrab/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/ftpgrab.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/ftpgrab/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/ftpgrab/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/ftpgrab/actions?workflow=build"><img src="https://img.shields.io/github/actions/workflow/status/crazy-max/ftpgrab/build.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build Status"></a>
  <a href="https://hub.docker.com/r/crazymax/ftpgrab/"><img src="https://img.shields.io/docker/stars/crazymax/ftpgrab.svg?style=flat-square&logo=docker" alt="Docker Stars"></a>
  <a href="https://hub.docker.com/r/crazymax/ftpgrab/"><img src="https://img.shields.io/docker/pulls/crazymax/ftpgrab.svg?style=flat-square&logo=docker" alt="Docker Pulls"></a>
  <br /><a href="https://goreportcard.com/report/github.com/crazy-max/ftpgrab"><img src="https://goreportcard.com/badge/github.com/crazy-max/ftpgrab?style=flat-square" alt="Go Report"></a>
  <a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>

---

## What is FTPGrab?

**FTPGrab** :zap: is a CLI tool for pulling files from remote FTP and SFTP servers to your NAS,
server, or computer. It is designed for repeatable, unattended transfers: define one or more
remote sources, run it on a schedule, filter by name or date, and keep track of what has already
been downloaded so later runs only fetch what is new. It can also send notifications when a job
completes, fails, or needs attention, which makes it easier to trust scheduled transfers.

It is a good fit for seedboxes, NAS setups, and small automation jobs where files need to move
reliably from remote storage to your own machine. FTPGrab is available as a
[single executable]({{ config.repo_url }}releases/latest) and as a [container image](install/docker.md).

## Features

* FTP and SFTP support
* Multiple remote sources per job
* Prevent re-downloads by tracking files that were already grabbed
* Efficient key/value store to audit downloaded files
* Built-in scheduling with cron expressions
* Include and exclude filters with regular expressions
* Date-based filtering
* Retry failed downloads
* Change file and folder permissions and ownership
* Preserve translated modification times on downloaded files
* Notifications through Mail, Slack, Webhook, scripts, and [more](config/index.md#reference)
* Enhanced logging
* Official [container image available](install/docker.md)

## License

This project is licensed under the terms of the MIT license.
