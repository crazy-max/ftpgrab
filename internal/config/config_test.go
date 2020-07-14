package config_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ftpgrab/ftpgrab/v7/internal/config"
	"github.com/ftpgrab/ftpgrab/v7/internal/model"
	"github.com/ftpgrab/ftpgrab/v7/pkg/utl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile(t *testing.T) {
	cases := []struct {
		name     string
		cfgfile  string
		wantData *config.Config
		wantErr  bool
	}{
		{
			name:    "Failed on non-existing file",
			cfgfile: "",
			wantErr: true,
		},
		{
			name:    "Fail on wrong file format",
			cfgfile: "./fixtures/config.invalid.yml",
			wantErr: true,
		},
		{
			name:    "Success",
			cfgfile: "./fixtures/config.test.yml",
			wantData: &config.Config{
				Db: (&model.Db{}).GetDefaults(),
				Server: &model.Server{
					FTP: &model.ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            utl.NewDuration(5 * time.Second),
						DisableEPSV:        utl.NewFalse(),
						TLS:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						LogTrace:           utl.NewFalse(),
					},
				},
				Download: &model.Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0644,
					ChmodDir:      0755,
					Retry:         3,
					HideSkipped:   utl.NewFalse(),
					CreateBaseDir: utl.NewFalse(),
				},
				Notif: &model.Notif{
					Mail: &model.NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						From:               "ftpgrab@example.com",
						To:                 "webmaster@example.com",
					},
					Slack: &model.NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
					},
					Webhook: &model.NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Method:   "GET",
						Headers: map[string]string{
							"content-type":  "application/json",
							"authorization": "Token123456",
						},
						Timeout: utl.NewDuration(10 * time.Second),
					},
				},
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(tt.cfgfile, "")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, cfg)
			if cfg != nil {
				assert.NotEmpty(t, cfg.String())
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	defer UnsetEnv("FTPGRAB_")

	testCases := []struct {
		desc     string
		cfgfile  string
		environ  []string
		expected interface{}
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
			expected: &config.Config{
				Db: (&model.Db{}).GetDefaults(),
				Server: &model.Server{
					FTP: &model.ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            utl.NewDuration(5 * time.Second),
						DisableEPSV:        utl.NewFalse(),
						TLS:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						LogTrace:           utl.NewFalse(),
					},
				},
				Download: &model.Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0644,
					ChmodDir:      0755,
					Retry:         3,
					HideSkipped:   utl.NewFalse(),
					CreateBaseDir: utl.NewFalse(),
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

			cfg, err := config.Load(tt.cfgfile, "")
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
		cfgfile  string
		environ  []string
		expected interface{}
		wantErr  bool
	}{
		{
			desc:    "env vars and invalid file",
			cfgfile: "./fixtures/config.invalid.yml",
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
			desc:    "ftp server (file) and notif mails (envs)",
			cfgfile: "./fixtures/config.ftp.yml",
			environ: []string{
				"FTPGRAB_NOTIF_MAIL_HOST=127.0.0.1",
				"FTPGRAB_NOTIF_MAIL_PORT=25",
				"FTPGRAB_NOTIF_MAIL_SSL=false",
				"FTPGRAB_NOTIF_MAIL_INSECURESKIPVERIFY=true",
				"FTPGRAB_NOTIF_MAIL_FROM=ftpgrab@foo.com",
				"FTPGRAB_NOTIF_MAIL_TO=webmaster@foo.com",
			},
			expected: &config.Config{
				Db: &model.Db{
					Path: "./fixtures/db/ftpgrab.db",
				},
				Server: &model.Server{
					FTP: &model.ServerFTP{
						Host:     "test.rebex.net",
						Port:     21,
						Username: "demo",
						Password: "password",
						Sources: []string{
							"/",
						},
						Timeout:            utl.NewDuration(5 * time.Second),
						DisableEPSV:        utl.NewFalse(),
						TLS:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewFalse(),
						LogTrace:           utl.NewFalse(),
					},
				},
				Download: &model.Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0644,
					ChmodDir:      0755,
					Retry:         3,
					HideSkipped:   utl.NewFalse(),
					CreateBaseDir: utl.NewFalse(),
				},
				Notif: &model.Notif{
					Mail: &model.NotifMail{
						Host:               "127.0.0.1",
						Port:               25,
						SSL:                utl.NewFalse(),
						InsecureSkipVerify: utl.NewTrue(),
						From:               "ftpgrab@foo.com",
						To:                 "webmaster@foo.com",
					},
				},
			},
			wantErr: false,
		},
		{
			desc:    "sftp server (file) and notif slack (envs)",
			cfgfile: "./fixtures/config.sftp.yml",
			environ: []string{
				"FTPGRAB_NOTIF_SLACK_WEBHOOKURL=https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
			},
			expected: &config.Config{
				Db: &model.Db{
					Path: "./fixtures/db/ftpgrab.db",
				},
				Server: &model.Server{
					SFTP: &model.ServerSFTP{
						Host:     "10.0.0.1",
						Port:     22,
						Username: "foo",
						Password: "bar",
						Sources: []string{
							"/",
						},
						Timeout:       utl.NewDuration(30 * time.Second),
						MaxPacketSize: 32768,
					},
				},
				Download: &model.Download{
					Output:        "./fixtures/downloads",
					UID:           os.Getuid(),
					GID:           os.Getgid(),
					ChmodFile:     0644,
					ChmodDir:      0755,
					Retry:         3,
					HideSkipped:   utl.NewTrue(),
					CreateBaseDir: utl.NewFalse(),
				},
				Notif: &model.Notif{
					Slack: &model.NotifSlack{
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

			cfg, err := config.Load(tt.cfgfile, "")
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidation(t *testing.T) {
	cases := []struct {
		name    string
		cfgfile string
	}{
		{
			name:    "Success",
			cfgfile: "./fixtures/config.validate.yml",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.Load(tt.cfgfile, "")
			require.NoError(t, err)

			//dec, err := env.Encode(cfg)
			//for _, value := range dec {
			//	fmt.Println(fmt.Sprintf(`%s=%s`, strings.Replace(value.Name, "TRAEFIK_", "FTPGRAB_", 1), value.Default))
			//}
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
