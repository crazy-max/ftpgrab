# Changelog

## 7.13.1 (2026/10/24)

This release is built with Go 1.26.10 to address [CVE-2026-111943](https://go.dev/issue/82345).
Certificate verification now rejects chains containing eight revoked eMudhra
intermediate certificates whose private keys were compromised.

The blocked intermediates are:

* Domain-validated TLS: `EM DV TLS CA - G3B` and `EM DV TLS CA - G3C`
* Organization-validated TLS: `EM OV TLS CA - G3B` and `EM OV TLS CA - G3C`
* Extended-validation TLS: `EM EV TLS CA - G3B` and `EM EV TLS CA - G3C`
* Code signing: `EM OV CS CA - G3B` and `EM OV CS CA - G3C`

The G3B intermediates belong to eMudhra BV, and the G3C intermediates belong to
eMudhra Inc. See the [eMudhra Intermediate CA Compromise](https://emudhra.we-have-a-problem.com/)
for more details.

There are no functional changes in FTPGrab itself compared to v7.13.0.

**Full Changelog**: [`v7.13.0...v7.13.1`](https://github.com/crazy-max/ftpgrab/compare/v7.13.0...v7.13.1)

## 7.13.0 (2026/04/17)

* Reconnect FTP sessions after timeout-poisoned transfers by @crazy-max in #508
* Fix webhook notification timeout context cleanup by @crazy-max in #513
* Remove outdated `GOMAXPROCS` override by @crazy-max in #514
* Log startup config as a structured debug field by @crazy-max in #509

**Full Changelog**: [`v7.12.0...v7.13.0`](https://github.com/crazy-max/ftpgrab/compare/v7.12.0...v7.13.0)

## 7.12.0 (2026/04/12)

* Explicit TLS support by @crazy-max in #477
* FTP path encoding option support for legacy servers by @crazy-max in #498
* Honor FTP and SFTP timeouts during stalled transfers by @crazy-max in #478
* Stage `tempFirst` downloads under a destination-local directory by @crazy-max in #481
* Close FTP and SFTP connections on init failures by @crazy-max in #480
* Make log file output follow the `log-json` setting by @crazy-max in #476
* Make shutdown safe before initialization by @crazy-max in #479
* Go 1.26 by @crazy-max in #497
* MkDocs Materials 9.7.5 by @crazy-max in #499
* Bump github.com/alecthomas/kong to 1.15.0 in #490
* Bump github.com/crazy-max/gonfig to 0.8.0 in #489
* Bump github.com/go-playground/validator/v10 to 10.30.2 in #445 #488
* Bump github.com/rs/zerolog to 1.35.0 in #485
* Bump golang.org/x/crypto to 0.49.0 in #464
* Bump golang.org/x/sys to 0.43.0 in #494

**Full Changelog**: [`v7.11.0...v7.12.0`](https://github.com/crazy-max/ftpgrab/compare/v7.11.0...v7.12.0)

## 7.11.0 (2025/12/24)

* Go 1.25 by @crazy-max in #443
* Alpine Linux 3.23 by @crazy-max in #444
* MkDocs Materials 9.6.20 by @crazy-max in #428
* Bump github.com/alecthomas/kong to 1.13.0 in #438
* Bump github.com/go-playground/validator/v10 to 10.30.0 in #440
* Bump github.com/pkg/sftp to 1.13.10 in #437
* Bump github.com/rs/zerolog to 1.34.0 in #422
* Bump github.com/stretchr/testify to 1.11.1 in #439
* Bump golang.org/x/crypto to 0.46.0 in #432
* Bump go.etcd.io/bbolt to 1.4.3 in #434

**Full Changelog**: [`v7.10.0...v7.11.0`](https://github.com/crazy-max/ftpgrab/compare/v7.10.0...v7.11.0)

## 7.10.0 (2024/12/20)

* Go 1.23 by @crazy-max in #397
* Alpine Linux 3.21 by @crazy-max in #397
* Bump github.com/alecthomas/kong to 1.6.0 in #323 #354 #393
* Bump github.com/go-playground/validator/v10 to 10.23.0 in #350 #370 #394
* Bump github.com/pkg/sftp to 1.13.7 in #326 #395
* Bump github.com/rs/zerolog to 1.33.0 in #344 #366
* Bump github.com/stretchr/testify to 1.10.0 in #351 #396
* Bump go.etcd.io/bbolt to 1.3.11 in #325 #348 #365 #377
* Bump golang.org/x/crypto to 0.31.0 in #324 #337 #352 #391
* Bump golang.org/x/net to 0.28.0 in #360 #353 #392

**Full Changelog**: [`v7.9.0...v7.10.0`](https://github.com/crazy-max/ftpgrab/compare/v7.9.0...v7.10.0)

## 7.9.0 (2023/12/16)

* Go 1.21 by @crazy-max in #322
* Alpine Linux 3.18 by @crazy-max in #322
* Bump github.com/alecthomas/kong to 0.8.0 in #306
* Bump github.com/crazy-max/gonfig to 0.7.0 in #291
* Bump github.com/jlaffaye/ftp to 0.2.0 by @crazy-max in #297 #301
* Bump github.com/go-playground/validator/v10 to 10.16.0 in #295 #321
* Bump github.com/rs/zerolog to 1.31.0 in #294 #320
* Bump github.com/stretchr/testify to 1.8.4 in #283 #302
* Bump golang.org/x/crypto to 0.8.0 in #293 #310
* Bump golang.org/x/net to 0.17.0 in #318
* Bump golang.org/x/sys to 0.15.0 in #287 #292 #319
* Bump go.etcd.io/bbolt to 1.3.7 in #277

**Full Changelog**: [`v7.8.0...v7.9.0`](https://github.com/crazy-max/ftpgrab/compare/v7.8.0...v7.9.0)

## 7.8.0 (2022/12/31)

* Option to escape all regular expression metacharacters by @crazy-max in #270
* Fix file mode type by @crazy-max in #269
* Move from `io/ioutil` to `io` and `os` packages by @crazy-max in #219
* Move `syscall` to `golang.org/x/sys` by @crazy-max in #220
* Go 1.19 by @crazy-max in #262 #253
* Alpine Linux 3.17 by @crazy-max in #268 #254 #223
* MkDocs Material 8.3.9 by @crazy-max in #256
* Enhance workflow by @crazy-max in #263 #218 #255
* Bump github.com/crazy-max/gonfig to 0.6.0 in #257
* Bump github.com/pkg/sftp to 1.13.5 in #208 #210 #246
* Bump github.com/rs/zerolog to 1.28.0 in #209 #211 #217 #245 #258
* Bump github.com/alecthomas/kong to 0.7.1 in #212 #215 #222 #230 #248 #266
* Bump github.com/go-playground/validator/v10 to 10.11.1 in #221 #229 #236 #261
* Bump github.com/stretchr/testify to 1.8.1 in #251 #264
* Bump github.com/docker/go-units to 0.5.0 in #259
* Bump golang.org/x/crypto to 0.4.0 by @crazy-max in #272
* Bump golang.org/x/sys to 0.3.0 by @crazy-max in #271

**Full Changelog**: [`v7.7.0...v7.8.0`](https://github.com/crazy-max/ftpgrab/compare/v7.7.0...v7.8.0)

## 7.7.0 (2021/09/05)

* Go 1.17 by @crazy-max in #203
* Wrong remaining time displayed by @crazy-max in #204
* Add `windows/arm64` artifact by @crazy-max in #205
* MkDocs Material 7.2.6
* Bump github.com/rs/zerolog to 1.24.0 in #207
* Bump github.com/crazy-max/gonfig to 0.5.0 in #206
* Bump github.com/gorilla/websocket to v1.4.2
* Bump github.com/go-playground/validator/v10 to 10.9.0 in #200 #202

**Full Changelog**: [`v7.6.0...v7.7.0`](https://github.com/crazy-max/ftpgrab/compare/v7.6.0...v7.7.0)

## 7.6.0 (2021/07/25)

* Add `linux/riscv64` artifact
* Alpine Linux 3.14
* MkDocs Materials 7.2.0
* GitHub Action cache backend by @crazy-max in #198
* Enhance issue template
* Bump github.com/pkg/sftp to 1.13.2 in #193 #196
* Bump github.com/go-playground/validator/v10 to 10.7.0 in #187 #195
* Bump go.etcd.io/bbolt to 1.3.6 in #190
* Bump github.com/rs/zerolog to 1.23.0 in #188 #194
* Bump github.com/alecthomas/kong to 0.2.17 in #191

**Full Changelog**: [`v7.5.0...v7.6.0`](https://github.com/crazy-max/ftpgrab/compare/v7.5.0...v7.6.0)

## 7.5.0 (2021/04/26)

* Add `disableMLSD` ftp option by @crazy-max in #176
* Fix Dockerfile

**Full Changelog**: [`v7.4.0...v7.5.0`](https://github.com/crazy-max/ftpgrab/compare/v7.4.0...v7.5.0)

## 7.4.0 (2021/04/25)

* Add `darwin/arm64` artifact by @crazy-max in #175
* Bump github.com/go-playground/validator/v10 to 10.5.0 in #171
* Use logger `PartsExclude` by @crazy-max in #174
* Go 1.16 by @crazy-max in #167
* Switch to goreleaser-xx by @crazy-max in #163
* MkDocs Materials 7.1.3
* Bump github.com/alecthomas/kong to 0.2.16 in #165
* Bump github.com/pkg/sftp to 1.13.0 in #164
* Bump github.com/rs/zerolog to 1.21.0 in #166

**Full Changelog**: [`v7.3.0...v7.4.0`](https://github.com/crazy-max/ftpgrab/compare/v7.3.0...v7.4.0)

## 7.3.0 (2021/02/19)

* Refactor CI and dev workflow with buildx bake by @crazy-max in #161
    * Add `image-local` target
    * Single job for artifacts and image
    * Add `armv5`, `ppc64le` and `s390x` artifacts
    * Upload artifacts
    * Validate
* Remove `linux/s390x` Docker platform support for now
* MkDocs Materials 6.2.8
* Bump github.com/stretchr/testify to 1.7.0 in #154
  Bump github.com/alecthomas/kong to 0.2.15 in #160

**Full Changelog**: [`v7.2.0...v7.3.0`](https://github.com/crazy-max/ftpgrab/compare/v7.2.0...v7.3.0)

## 7.2.0 (2020/11/29)

* Allow downloading files to a temp dir first by @crazy-max in #149
* Allow disabling log timestamp by @crazy-max in #148
* Add script notification by @crazy-max in #147
* Bump github.com/crazy-max/gonfig to 0.4.0 in #140

**Full Changelog**: [`v7.1.1...v7.2.0`](https://github.com/crazy-max/ftpgrab/compare/v7.1.1...v7.2.0)

## 7.1.1 (2020/11/02)

* Use embedded tzdata package
* Remove `--timezone` flag
* Docker image also available on [GitHub Container Registry](https://github.com/users/crazy-max/packages/container/package/ftpgrab)
* Use Docker meta action to handle tags and labels

**Full Changelog**: [`v7.1.0...v7.1.1`](https://github.com/crazy-max/ftpgrab/compare/v7.1.0...v7.1.1)

## 7.1.0 (2020/10/04)

* Allow disabling `OPTS UTF8 ON` command
* Refactor to start working on #48
* Switch to Docker actions
* Go 1.15
* Update `GOPROXY` setting
* Update deps

**Full Changelog**: [`v7.0.1...v7.1.0`](https://github.com/crazy-max/ftpgrab/compare/v7.0.1...v7.1.0)

## 7.0.1 (2020/08/04)

* Fix SFTP not taken into account

**Full Changelog**: [`v7.0.0...v7.0.1`](https://github.com/crazy-max/ftpgrab/compare/v7.0.0...v7.0.1)

## 7.0.0 (2020/07/18)

:warning: See **Migration notes** in the documentation for breaking changes.

* Repository moved to [crazy-max/ftpgrab](https://github.com/crazy-max/ftpgrab)
* DockerHub repository moved to [crazymax/ftpgrab](https://hub.docker.com/r/crazymax/ftpgrab)
* Configuration transposed into environment variables by @crazy-max in #90
* `FTPGRAB_DB` env var renamed `FTPGRAB_DB_PATH`
* `key` field for SFTP authentication has been renamed `keyFile`
* Add `keyPassphrase` to provide a passphrase linked to `keyFile`
* Improve configuration validation
* All fields in configuration now _camelCased_
* Add tests and coverage
* Seek configuration file from default places
* Configuration file not required anymore
* Switch to [gonfig](https://github.com/crazy-max/gonfig)
* Add fields to load sensitive values from file
* Update deps

**Full Changelog**: [`v6.5.0...v7.0.0`](https://github.com/crazy-max/ftpgrab/compare/v6.5.0...v7.0.0)

## 6.5.0 (2020/07/07)

* Docs website with mkdocs
* Move documentation to main repository
* Update deps

**Full Changelog**: [`v6.4.0...v6.5.0`](https://github.com/crazy-max/ftpgrab/compare/v6.4.0...v6.5.0)

## 6.4.0 (2020/05/17)

* Use kong command-line parser
* Switch to Open Container Specification labels as label-schema.org ones are deprecated
* Update deps

**Full Changelog**: [`v6.3.0...v6.4.0`](https://github.com/crazy-max/ftpgrab/compare/v6.3.0...v6.4.0)

## 6.3.0 (2020/01/19)

* Only accept duration as timeout value for FTP, SFTP and Webhook notif config in #69
* Update [pkg/sftp](https://github.com/pkg/sftp) module

**Full Changelog**: [`v6.2.0...v6.3.0`](https://github.com/crazy-max/ftpgrab/compare/v6.2.0...v6.3.0)

## 6.2.0 (2019/12/19)

* Add Slack notifier
* Update deps
* Go 1.13.5
* Seconds field optional for schedule

**Full Changelog**: [`v6.1.0...v6.2.0`](https://github.com/crazy-max/ftpgrab/compare/v6.1.0...v6.2.0)

## 6.1.0 (2019/10/13)

* Multi-platform Docker image
* Move [ftpgrab/docker](https://github.com/ftpgrab/docker) repo here
* Go 1.12.10
* Use GOPROXY
* Stop publishing Docker image on Quay
* Switch to GitHub Actions
* Add instructions to create a Linux service
* Remove `--docker` flag
* Allow overriding database path through `FTPGRAB_DB` env var
* Allow overriding download output path through `FTPGRAB_DOWNLOAD_OUTPUT` env var

**Full Changelog**: [`v6.0.2...v6.1.0`](https://github.com/crazy-max/ftpgrab/compare/v6.0.2...v6.1.0)

## 6.0.2 (2019/07/27)

* Use `io.Copy` to avoid crash due to insufficient memory

**Full Changelog**: [`v6.0.1...v6.0.2`](https://github.com/crazy-max/ftpgrab/compare/v6.0.1...v6.0.2)

## 6.0.1 (2019/07/24)

* Fix cron stopped after first trigger

**Full Changelog**: [`v6.0.0...v6.0.1`](https://github.com/crazy-max/ftpgrab/compare/v6.0.0...v6.0.1)

## 6.0.0 (2019/07/21)

:warning: See **Migration notes** in the documentation for breaking changes.

* Log skip status
* Set ServerName field if implicit TLS
* Switch to [jlaffaye/ftp](https://github.com/jlaffaye/ftp) module
    * Fix race condition
    * Performance improvement

**Full Changelog**: [`v5.5.0...v6.0.0`](https://github.com/crazy-max/ftpgrab/compare/v5.5.0...v6.0.0)

## 5.5.0 (2019/07/20)

* Switch to [crazy-max/goftp](https://github.com/crazy-max/goftp) in #55

**Full Changelog**: [`v5.4.1...v5.5.0`](https://github.com/crazy-max/ftpgrab/compare/v5.4.1...v5.5.0)

## 5.4.1 (2019/07/18)

* Fix durafmt runtime error

**Full Changelog**: [`v5.4.0...v5.4.1`](https://github.com/crazy-max/ftpgrab/compare/v5.4.0...v5.4.1)

## 5.4.0 (2019/07/18)

* Improve logging
* Display next execution time
* Use v3 robfig/cron
* Always run on startup
* Go 1.12.4

**Full Changelog**: [`v5.3.0...v5.4.0`](https://github.com/crazy-max/ftpgrab/compare/v5.3.0...v5.4.0)

## 5.3.0 (2019/05/04)

* Escape all regexp metacharacters on read dir in #49
* Remove unused field
* Go 1.12
* Update deps

**Full Changelog**: [`v5.2.0...v5.3.0`](https://github.com/crazy-max/ftpgrab/compare/v5.2.0...v5.3.0)

## 5.2.0 (2019/03/29)

:warning: See **Migration notes** in the documentation for breaking changes.

* Add webhook notification method
* Remove unnecessary `connections_per_host` field in #48
* Fix log folder creation

**Full Changelog**: [`v5.1.1...v5.2.0`](https://github.com/crazy-max/ftpgrab/compare/v5.1.1...v5.2.0)

## 5.1.1 (2019/02/18)

* Blackfriday module fixed through hermes v2.0.2 (matcornic/hermes#51)

**Full Changelog**: [`v5.1.0...v5.1.1`](https://github.com/crazy-max/ftpgrab/compare/v5.1.0...v5.1.1)

## 5.1.0 (2019/02/14)

:warning: See **Migration notes** in the documentation for breaking changes.

* Add SFTP support in #42

**Full Changelog**: [`v5.0.1...v5.1.0`](https://github.com/crazy-max/ftpgrab/compare/v5.0.1...v5.1.0)

## 5.0.1 (2019/02/13)

* Fix high CPU load on schedule
* Add support for FreeBSD

**Full Changelog**: [`v5.0.0...v5.0.1`](https://github.com/crazy-max/ftpgrab/compare/v5.0.0...v5.0.1)

## 5.0.0 (2019/02/12)

:warning: See **Migration notes** in the documentation for breaking changes.

* BIG rewrite in #36
* Multiplatform : Linux, macOS and Windows on architectures like amd64, 386, ARM and others
* Modern CLI interactions
* Yaml Configuration file
* Detect and merge configuration
* Handle defaults
* Add [Goreleaser](https://goreleaser.com/)
* [Bolt](https://github.com/etcd-io/bbolt) db to audit files already downloaded
* Native FTP client
* Logging with [zerolog](https://github.com/rs/zerolog)
* Send reports through email
* Generate responsive and beautiful email reports through [hermes](https://github.com/matcornic/hermes/)
* Lightweight Docker image (~6MB)
* Docker image moved to a dedicated organization on [Docker Hub](https://hub.docker.com/u/ftpgrab) and [Quay](https://quay.io/organization/ftpgrab).
* [Embedded cron](https://github.com/crazy-max/cron) using go routines
* Manage base dir
* Set original modtime
* Include/exclude based on regexp
* Ignore files by date in #39
* Handle mutex

**Full Changelog**: [`v4.3.5...v5.0.0`](https://github.com/crazy-max/ftpgrab/compare/v4.3.5...v5.0.0)

## 4.3.5 (2019/02/04)

* Switch to Travis CI (com)

**Full Changelog**: [`v4.3.4...v4.3.5`](https://github.com/crazy-max/ftpgrab/compare/v4.3.4...v4.3.5)

## 4.3.4 (2018/08/15)

* Empty folder leeds to spinlock in #33

**Full Changelog**: [`v4.3.3...v4.3.4`](https://github.com/crazy-max/ftpgrab/compare/v4.3.3...v4.3.4)

## 4.3.3 (2018/05/14)

* nawk and gawk not required anymore in #38

**Full Changelog**: [`v4.3.2...v4.3.3`](https://github.com/crazy-max/ftpgrab/compare/v4.3.2...v4.3.3)

## 4.3.2 (2018/04/20)

* Detect if file size is currently changing and hold for download in #37

**Full Changelog**: [`v4.3.1...v4.3.2`](https://github.com/crazy-max/ftpgrab/compare/v4.3.1...v4.3.2)

## 4.3.1 (2018/01/15)

* Fix issue while checking source hash in #35

**Full Changelog**: [`v4.3.0...v4.3.1`](https://github.com/crazy-max/ftpgrab/compare/v4.3.0...v4.3.1)

## 4.3.0 (2017/12/26)

* Add an exclude filter for files through `DL_EXCLUDE_REGEX` in #27

**Full Changelog**: [`v4.2.4...v4.3.0`](https://github.com/crazy-max/ftpgrab/compare/v4.2.4...v4.3.0)

## 4.2.4 (2017/11/01)

* Do not exit if connection failed

**Full Changelog**: [`v4.2.3...v4.2.4`](https://github.com/crazy-max/ftpgrab/compare/v4.2.3...v4.2.4)

## 4.2.3 (2017/10/30)

* Fix files download again in #32

**Full Changelog**: [`v4.2.2...v4.2.3`](https://github.com/crazy-max/ftpgrab/compare/v4.2.2...v4.2.3)

## 4.2.2 (2017/10/29)

* Rebuild PATH

**Full Changelog**: [`v4.2.1...v4.2.2`](https://github.com/crazy-max/ftpgrab/compare/v4.2.1...v4.2.2)

## 4.2.1 (2017/10/16)

* Add ssmtp on Docker image to send emails
* Use sendmail instead of mail command

**Full Changelog**: [`v4.2.0...v4.2.1`](https://github.com/crazy-max/ftpgrab/compare/v4.2.0...v4.2.1)

## 4.2.0 (2017/10/15)

:warning: See **Migration notes** in the documentation for breaking changes.

* Add Docker image (more info on [docker repository](https://github.com/ftpgrab/docker))
* Remove init script
* Fix issue while resuming downloads
* Move script to `/usr/bin`
* Coding style

**Full Changelog**: [`v4.1.1...v4.2.0`](https://github.com/crazy-max/ftpgrab/compare/v4.1.1...v4.2.0)

## 4.1.1 (2017/04/26)

* Add tests in #30
* Use type instead of which in #29
* Fix error prone and performance issues
* Coding style
* Add default config
* Add Codacy

**Full Changelog**: [`v4.1...v4.1.1`](https://github.com/crazy-max/ftpgrab/compare/v4.1...v4.1.1)

## 4.1 (2017/03/15)

:warning: See **Migration notes** in the documentation for breaking changes.

* Rename the project ftpgrab ! in #28

**Full Changelog**: [`v4.0...v4.1`](https://github.com/crazy-max/ftpgrab/compare/v4.0...v4.1)

## 4.0 (2017/03/14)

:warning: See **Migration notes** in the documentation for breaking changes.

* Shuffle file/folder listing by @bwibwi13 in #25
* Allow multiple instances in #22

**Full Changelog**: [`v3.2...v4.0`](https://github.com/crazy-max/ftpgrab/compare/v3.2...v4.0)

## 3.2 (2016/06/20)

* Add messages for permission issue in #19
* Move some instructions to Wiki in #18
* MIT License

**Full Changelog**: [`v3.1...v3.2`](https://github.com/crazy-max/ftpgrab/compare/v3.1...v3.2)

## 3.1 (2016/03/27)

**You have to edit the config file `ftp-sync.conf` if you upgrade from a previous release!**

* Add multiple ftp sources paths in #18
* Sed not escaping `&` char in #17
* Add `DL_CREATE_BASEDIR` option to create basename of a ftp source path in the destination folder.

**Full Changelog**: [`v3.0...v3.1`](https://github.com/crazy-max/ftpgrab/compare/v3.0...v3.1)

## 3.0 (2016/03/20)

**You have to edit the config file `ftp-sync.conf` if you upgrade from a previous release!**

* MD5 file not created with text mode in #16
* Implement FTPS support for Curl in #15
* Implement resume downloads support in #14
* Add DEBUG option
* Full Curl implementation when selected for file size and list files
* Bug with ftpsyncGetHumanSize function
* Display download regex
* Add sha1 hash type
* Bug with special chars for curl method
* Bug with bash condition

**Full Changelog**: [`v2.03...v3.0`](https://github.com/crazy-max/ftpgrab/compare/v2.03...v3.0)

## 2.03 (2015/03/22)

* Change location of MD5 file

**Full Changelog**: [`v2.02...v2.03`](https://github.com/crazy-max/ftpgrab/compare/v2.02...v2.03)

## 2.02 (2015/03/21)

* Bug checking MD5 in #11

**Full Changelog**: [`v2.01...v2.02`](https://github.com/crazy-max/ftpgrab/compare/v2.01...v2.02)

## 2.01 (2015/03/20)

* Bug download with sqlite3 in #10

**Full Changelog**: [`v2.00...v2.01`](https://github.com/crazy-max/ftpgrab/compare/v2.00...v2.01)

## 2.00 (2015/03/19)

* Add SQLite method to store MD5 hash in #8

**Full Changelog**: [`v1.95...v2.00`](https://github.com/crazy-max/ftpgrab/compare/v1.95...v2.00)

## 1.95 (2014/08/09)

* Bug trailing slash in #6

**Full Changelog**: [`v1.94...v1.95`](https://github.com/crazy-max/ftpgrab/compare/v1.94...v1.95)

## 1.94 (2014/05/22)

* Bug replacing destination folder

**Full Changelog**: [`v1.93...v1.94`](https://github.com/crazy-max/ftpgrab/compare/v1.93...v1.94)

## 1.93 (2014/02/16)

* Adding hide progress option

**Full Changelog**: [`v1.92...v1.93`](https://github.com/crazy-max/ftpgrab/compare/v1.92...v1.93)

## 1.92 (2013/12/01)

* Bug with the config file

**Full Changelog**: [`v1.91...v1.92`](https://github.com/crazy-max/ftpgrab/compare/v1.91...v1.92)

## 1.91 (2013/12/01)

* Adding curl download method

**Full Changelog**: [`v1.9...v1.91`](https://github.com/crazy-max/ftpgrab/compare/v1.9...v1.91)

## 1.9 (2013/10/30)

* Remove progress filter on wget

**Full Changelog**: [`v1.8...v1.9`](https://github.com/crazy-max/ftpgrab/compare/v1.8...v1.9)

## 1.8 (2013/10/12)

* Bug with empty folders

**Full Changelog**: [`v1.7...v1.8`](https://github.com/crazy-max/ftpgrab/compare/v1.7...v1.8)

## 1.7 (2013/10/06)

* Adding external config file
* Add gawk as required package
* Update README.md with awk problem
* Change perms recursively when downloads are finished

**Full Changelog**: [`v1.6...v1.7`](https://github.com/crazy-max/ftpgrab/compare/v1.6...v1.7)

## 1.6 (2013/07/10)

* Misspelling
* Decoding wget problem
* Alternative to kill old and sub process

**Full Changelog**: [`v1.5...v1.6`](https://github.com/crazy-max/ftpgrab/compare/v1.5...v1.6)

## 1.5 (2013/06/10)

* Add synology example

**Full Changelog**: [`v1.4...v1.5`](https://github.com/crazy-max/ftpgrab/compare/v1.4...v1.5)

## 1.4 (2013/06/05)

* Check process already running

**Full Changelog**: [`v1.3...v1.4`](https://github.com/crazy-max/ftpgrab/compare/v1.3...v1.4)

## 1.3 (2013/06/02)

* Use wget instead of curlftpfs

**Full Changelog**: [`v1.2...v1.3`](https://github.com/crazy-max/ftpgrab/compare/v1.2...v1.3)

## 1.2 (2013/06/01)

* Adding email var to receive logs

**Full Changelog**: [`v1.1...v1.2`](https://github.com/crazy-max/ftpgrab/compare/v1.1...v1.2)

## 1.1 (2013/05/31)

* Remove dualEcho
* Improvement of the error log with exec and tail
* Change MD5 filter
* Filter bug and add grep search for hash

**Full Changelog**: [`v1.0...v1.1`](https://github.com/crazy-max/ftpgrab/compare/v1.0...v1.1)

## 1.0 (2013/05/24)

* Initial version
