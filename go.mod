module github.com/ftpgrab/ftpgrab

go 1.13

require (
	github.com/alecthomas/kong v0.2.9
	github.com/containous/traefik/v2 v2.2.3
	github.com/docker/go-units v0.4.0
	github.com/go-gomail/gomail v0.0.0-20160411212932-81ebce5c23df
	github.com/go-playground/validator/v10 v10.3.0
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4
	github.com/ilya1st/rotatewriter v0.0.0-20171126183947-3df0c1a3ed6d
	github.com/imdario/mergo v0.3.9
	github.com/jlaffaye/ftp v0.0.0-20190828173736-6aaa91c7796e
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/nlopes/slack v0.6.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.19.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20200429183012-4b2356b1ed79
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

// Containous forks
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
)
