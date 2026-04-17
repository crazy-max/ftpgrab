package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/gonfig/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile(t *testing.T) {
	cases := []struct {
		name     string
		cli      Cli
		wantData *Config
		wantErr  bool
	}{
		{
			name:    "Failed on non-existing file",
			wantErr: true,
		},
		{
			name: "Fail on wrong file format",
			cli: Cli{
				Cfgfile: "./fixtures/config.invalid.yml",
			},
			wantErr: true,
		},
		{
			name: "Success",
			cli: Cli{
				Cfgfile: "./fixtures/config.test.yml",
			},
			wantData: &Config{
				Cli: Cli{
					Cfgfile: "./fixtures/config.test.yml",
				},
				Db: (&Db{}).GetDefaults(),
				Server: &Server{
					FTP: &ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            new(5 * time.Second),
						PathEncoding:       "utf-8",
						DisableUTF8:        new(false),
						DisableEPSV:        new(false),
						DisableMLSD:        new(false),
						EscapeRegexpMeta:   new(false),
						TLS:                new(false),
						ExplicitTLS:        new(false),
						InsecureSkipVerify: new(false),
						LogTrace:           new(false),
					},
				},
				Download: &Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0o644,
					ChmodDir:      0o755,
					Since:         "2019-02-01T18:50:05Z",
					SinceTime:     time.Date(2019, 2, 1, 18, 50, 05, 0, time.UTC),
					Retry:         3,
					HideSkipped:   new(false),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
				Notif: &Notif{
					Mail: &NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                new(false),
						InsecureSkipVerify: new(false),
						From:               "ftpgrab@example.com",
						To:                 "webmaster@example.com",
					},
					Script: &NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
					Slack: &NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Webhook: &NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "application/json",
							"authorization": "Token123456",
						},
						Timeout: new(10 * time.Second),
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.cli, Meta{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, cfg)
		})
	}
}

func TestLoadEnv(t *testing.T) {
	defer UnsetEnv("FTPGRAB_")

	testCases := []struct {
		desc     string
		cli      Cli
		environ  []string
		expected any
		wantErr  bool
	}{
		{
			desc:     "no env vars",
			environ:  nil,
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "ftp server",
			environ: []string{
				"FTPGRAB_SERVER_FTP_HOST=test.rebex.net",
				"FTPGRAB_SERVER_FTP_USERNAME=demo",
				"FTPGRAB_SERVER_FTP_PASSWORD=password",
				"FTPGRAB_SERVER_FTP_SOURCES=/",
				"FTPGRAB_DOWNLOAD_OUTPUT=./fixtures/downloads",
			},
			expected: &Config{
				Db: (&Db{}).GetDefaults(),
				Server: &Server{
					FTP: &ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            new(5 * time.Second),
						PathEncoding:       "utf-8",
						DisableUTF8:        new(false),
						DisableEPSV:        new(false),
						DisableMLSD:        new(false),
						EscapeRegexpMeta:   new(false),
						TLS:                new(false),
						ExplicitTLS:        new(false),
						InsecureSkipVerify: new(false),
						LogTrace:           new(false),
					},
				},
				Download: &Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0o644,
					ChmodDir:      0o755,
					Retry:         3,
					HideSkipped:   new(false),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
			},
			wantErr: false,
		},
		{
			desc: "sftp server",
			environ: []string{
				"FTPGRAB_SERVER_SFTP_HOST=10.0.0.1",
				"FTPGRAB_SERVER_SFTP_USERNAMEFILE=./fixtures/run_secrets_username",
				"FTPGRAB_SERVER_SFTP_PASSWORDFILE=./fixtures/run_secrets_password",
				"FTPGRAB_SERVER_SFTP_SOURCES=/",
				"FTPGRAB_DOWNLOAD_OUTPUT=./fixtures/downloads",
			},
			expected: &Config{
				Db: (&Db{}).GetDefaults(),
				Server: &Server{
					//nolint:gosec // Test fixture paths below point to secret files, they are not embedded credentials.
					SFTP: &ServerSFTP{
						Host:         "10.0.0.1",
						Port:         22,
						UsernameFile: "./fixtures/run_secrets_username",
						PasswordFile: "./fixtures/run_secrets_password",
						Sources: []string{
							"/",
						},
						Timeout:       new(30 * time.Second),
						MaxPacketSize: 32768,
					},
				},
				Download: &Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0o644,
					ChmodDir:      0o755,
					Retry:         3,
					HideSkipped:   new(false),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
			},
			wantErr: false,
		},
		{
			desc: "ftp and sftp server defined",
			environ: []string{
				"FTPGRAB_SERVER_FTP_HOST=test.rebex.net",
				"FTPGRAB_SERVER_FTP_USERNAME=demo",
				"FTPGRAB_SERVER_FTP_PASSWORD=password",
				"FTPGRAB_SERVER_FTP_SOURCES=/",
				"FTPGRAB_SERVER_SFTP_HOST=10.0.0.1",
				"FTPGRAB_SERVER_SFTP_PORT=22",
				"FTPGRAB_SERVER_SFTP_USERNAME=foo",
				"FTPGRAB_SERVER_SFTP_PASSWORD=bar",
				"FTPGRAB_SERVER_SFTP_SOURCES=/",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "ftp implicit and explicit tls enabled",
			environ: []string{
				"FTPGRAB_SERVER_FTP_HOST=test.rebex.net",
				"FTPGRAB_SERVER_FTP_USERNAME=demo",
				"FTPGRAB_SERVER_FTP_PASSWORD=password",
				"FTPGRAB_SERVER_FTP_SOURCES=/",
				"FTPGRAB_SERVER_FTP_TLS=true",
				"FTPGRAB_SERVER_FTP_EXPLICITTLS=true",
				"FTPGRAB_DOWNLOAD_OUTPUT=./fixtures/downloads",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			UnsetEnv("FTPGRAB_")

			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					os.Setenv(n[0], n[1])
				}
			}

			cfg, err := Load(tt.cli, Meta{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestLoadMixed(t *testing.T) {
	defer UnsetEnv("FTPGRAB_")

	testCases := []struct {
		desc     string
		cli      Cli
		environ  []string
		expected any
		wantErr  bool
	}{
		{
			desc: "env vars and invalid file",
			cli: Cli{
				Cfgfile: "./fixtures/config.invalid.yml",
			},
			environ: []string{
				"FTPGRAB_SERVER_FTP_HOST=test.rebex.net",
				"FTPGRAB_SERVER_FTP_USERNAME=demo",
				"FTPGRAB_SERVER_FTP_PASSWORD=password",
				"FTPGRAB_SERVER_FTP_SOURCES=/",
				"FTPGRAB_DOWNLOAD_OUTPUT=./fixtures/downloads",
			},
			expected: nil,
			wantErr:  true,
		},
		{
			desc: "ftp server (file) and notif mails (envs)",
			cli: Cli{
				Cfgfile: "./fixtures/config.ftp.yml",
			},
			environ: []string{
				"FTPGRAB_NOTIF_MAIL_HOST=127.0.0.1",
				"FTPGRAB_NOTIF_MAIL_PORT=25",
				"FTPGRAB_NOTIF_MAIL_SSL=false",
				"FTPGRAB_NOTIF_MAIL_INSECURESKIPVERIFY=true",
				"FTPGRAB_NOTIF_MAIL_FROM=ftpgrab@foo.com",
				"FTPGRAB_NOTIF_MAIL_TO=webmaster@foo.com",
			},
			expected: &Config{
				Cli: Cli{
					Cfgfile: "./fixtures/config.ftp.yml",
				},
				Db: &Db{
					Path: "./fixtures/db/ftpgrab.db",
				},
				Server: &Server{
					FTP: &ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            new(5 * time.Second),
						PathEncoding:       "utf-8",
						DisableUTF8:        new(false),
						DisableEPSV:        new(false),
						DisableMLSD:        new(false),
						EscapeRegexpMeta:   new(false),
						TLS:                new(false),
						ExplicitTLS:        new(false),
						InsecureSkipVerify: new(false),
						LogTrace:           new(false),
					},
				},
				Download: &Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0o644,
					ChmodDir:      0o755,
					Retry:         3,
					HideSkipped:   new(false),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
				Notif: &Notif{
					Mail: &NotifMail{
						Host:               "127.0.0.1",
						Port:               25,
						SSL:                new(false),
						InsecureSkipVerify: new(true),
						From:               "ftpgrab@foo.com",
						To:                 "webmaster@foo.com",
					},
				},
			},
			wantErr: false,
		},
		{
			desc: "sftp server (file) and notif slack (envs)",
			cli: Cli{
				Cfgfile: "./fixtures/config.sftp.yml",
			},
			environ: []string{
				"FTPGRAB_NOTIF_SLACK_WEBHOOKURL=https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
			},
			expected: &Config{
				Cli: Cli{
					Cfgfile: "./fixtures/config.sftp.yml",
				},
				Db: &Db{
					Path: "./fixtures/db/ftpgrab.db",
				},
				Server: &Server{
					SFTP: &ServerSFTP{
						Host:     "10.0.0.1",
						Port:     22,
						Username: "foo",
						Password: "bar",
						Sources: []string{
							"/",
						},
						Timeout:       new(30 * time.Second),
						MaxPacketSize: 32768,
					},
				},
				Download: &Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0o644,
					ChmodDir:      0o755,
					Retry:         3,
					HideSkipped:   new(true),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
				Notif: &Notif{
					Slack: &NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			UnsetEnv("FTPGRAB_")

			if tt.environ != nil {
				for _, environ := range tt.environ {
					n := strings.SplitN(environ, "=", 2)
					os.Setenv(n[0], n[1])
				}
			}

			cfg, err := Load(tt.cli, Meta{})
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestNormalizeFTPPathEncoding(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "default", value: "", want: "utf-8"},
		{name: "utf8 alias", value: "utf8", want: "utf-8"},
		{name: "cp1251 alias", value: "cp1251", want: "windows-1251"},
		{name: "windows1251 alias", value: "windows1251", want: "windows-1251"},
		{name: "invalid", value: "koi8-r", wantErr: true},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeFTPPathEncoding(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidation(t *testing.T) {
	cases := []struct {
		name string
		cli  Cli
	}{
		{
			name: "Success",
			cli: Cli{
				Cfgfile: "./fixtures/config.validate.yml",
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.cli, Meta{})
			require.NoError(t, err)

			dec, err := env.Encode("FTPGRAB_", cfg)
			require.NoError(t, err)
			for _, value := range dec {
				fmt.Printf(`%s=%s\n`, value.Name, value.Default)
			}
		})
	}
}

func UnsetEnv(prefix string) (restore func()) {
	before := map[string]string{}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]

		os.Unsetenv(parts[0])
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}

			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			// Check if the envar previously existed
			v, ok := before[parts[0]]
			if !ok {
				// This is a newly added envar with prefix, zap it
				os.Unsetenv(parts[0])
				continue
			}

			if parts[1] != v {
				// If the envar value has changed, set it back
				os.Setenv(parts[0], v)
			}
		}

		// Still need to check if there have been any deleted envars
		for k, v := range before {
			if _, ok := after[k]; !ok {
				// k is not present in after, so we set it.
				os.Setenv(k, v)
			}
		}
	}
}
