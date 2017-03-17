# Changelog

## 4.1 (2017/03/15)

* Rename the project ftpgrab ! (Issue #28)

## 4.0 (2017/03/14)

* Shuffle file/folder listing (Issue #25)
* Allow multiple instances (Issue #22)

## 3.2 (2016/06/20)

* Add messages for permission issue (Issue #19)
* Move some instructions to Wiki (Issue #18)
* Update `ISSUE_TEMPLATE.md`
* Add [.editorconfig](http://editorconfig.org/)
* MIT License

## 3.1 (2016/03/27)

**You have to edit the config file `ftp-sync.conf` if you upgrade from a previous release!**

* Add multiple ftp sources paths (Issue #18)
* Sed not escaping & char (Issue #17)
* Add `DL_CREATE_BASEDIR` option to create basename of a ftp source path in the destination folder.

## 3.0 (2016/03/20)

**You have to edit the config file `ftp-sync.conf` if you upgrade from a previous release!**

* MD5 file not created with text mode (Issue #16)
* Implement FTPS support for Curl (Issue #15)
* Implement resume downloads support (Issue #14)
* Add DEBUG option
* Full Curl implementation when selected for file size and list files
* Bug with ftpsyncGetHumanSize function
* Display download regex
* Add sha1 hash type
* Bug with special chars for curl method
* Bug with bash condition
* Add `Found a bug?` section in README.md
* Add `ISSUE_TEMPLATE.md`

## 2.03 (2015/03/22)

* Change location of MD5 file

## 2.02 (2015/03/21)

* Bug checking MD5 (Issue #11)

## 2.01 (2015/03/20)

* Bug download with sqlite3 (Issue #10)

## 2.00 (2015/03/19)

* Add SQLite method to store MD5 hash (Issue #8)

## 1.95 (2014/08/09)

* Bug trailing slash  (Issue #6)

## 1.94 (2014/05/22)

* Bug replacing destination folder

## 1.93 (2014/02/16)

* Update README.md and .gitignore
* New year!
* Adding hide progress option

## 1.92 (2013/12/01)

* Bug with the config file

## 1.91 (2013/12/01)

* Adding curl download method

## 1.9 (2013/10/30)

* Remove progress filter on wget

## 1.8 (2013/10/12)

* Update README.md
* Bug with empty folders

## 1.7 (2013/10/06)

* Adding external config file
* Add gawk as required package
* Update README.md with awk problem
* Change perms recursively when downloads are finished

## 1.6 (2013/07/10)

* Misspelling
* Decoding wget problem
* Alternative to kill old and sub process

## 1.5 (2013/06/10)

* Update README.md
* Add synology example

## 1.4 (2013/06/05)

* Check process already running

## 1.3 (2013/06/02)

* Use wget instead of curlftpfs

## 1.2 (2013/06/01)

* Adding email var to receive logs

## 1.1 (2013/05/31)

* Remove dualEcho
* Improvement of the error log with exec and tail
* Change MD5 filter
* Filter bug and add grep search for hash

## 1.0 (2013/05/24)

* Update README.md
* Initial version
