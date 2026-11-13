package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/gonfig/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadConcurrency(t *testing.T) {
	for _, protocol := range []string{"ftp", "sftp"} {
		for _, source := range []string{"file", "env"} {
			for _, tc := range []struct {
				name    string
				value   string
				want    int
				wantErr bool
			}{
				{
					name: "default",
					want: 1,
				},
				{
					name:  "parallel",
					value: "4",
					want:  4,
				},
				{
					name:    "zero",
					value:   "0",
					wantErr: true,
				},
				{
					name:    "negative",
					value:   "-1",
					wantErr: true,
				},
			} {
				t.Run(protocol+"/"+source+"/"+tc.name, func(t *testing.T) {
					filename := "./fixtures/config." + protocol + ".yml"
					if tc.value != "" {
						if source == "env" {
							t.Setenv("FTPGRAB_DOWNLOAD_CONCURRENCY", tc.value)
						} else {
							content, err := os.ReadFile(filename)
							require.NoError(t, err)
							content = bytes.Replace(content, []byte("download:"), []byte("download:\n  concurrency: "+tc.value), 1)
							filename = filepath.Join(t.TempDir(), "ftpgrab.yml")
							// #nosec G703 -- The path is a test-owned temporary directory plus a fixed filename.
							require.NoError(t, os.WriteFile(filename, content, 0o600))
						}
					}
					cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
					if tc.wantErr {
						require.ErrorContains(t, err, "Concurrency")
						return
					}
					require.NoError(t, err)
					assert.Equal(t, tc.want, cfg.Download.Concurrency)
				})
			}
		}
	}
}

func TestRemovedEscapeRegexpMeta(t *testing.T) {
	for _, source := range []string{"file", "env", "unset"} {
		for _, value := range []string{"true", "false"} {
			t.Run(source+"/"+value, func(t *testing.T) {
				var output bytes.Buffer
				previous := log.Logger
				log.Logger = zerolog.New(&output)
				t.Cleanup(func() { log.Logger = previous })
				filename := "./fixtures/config.ftp.yml"
				switch source {
				case "file":
					content, err := os.ReadFile(filename)
					require.NoError(t, err)
					content = bytes.Replace(content, []byte("  ftp:"), []byte("  ftp:\n    escapeRegexpMeta: "+value), 1)
					filename = filepath.Join(t.TempDir(), "ftpgrab.yml")
					// #nosec G703 -- The path is a test-owned temporary directory plus a fixed filename.
					require.NoError(t, os.WriteFile(filename, content, 0o600))
				case "env":
					t.Setenv("FTPGRAB_SERVER_FTP_ESCAPEREGEXPMETA", value)
				}
				cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
				require.NoError(t, err)
				if source == "unset" {
					assert.Nil(t, cfg.Server.FTP.EscapeRegexpMeta)
					assert.NotContains(t, output.String(), "was removed and is ignored")
				} else {
					require.NotNil(t, cfg.Server.FTP.EscapeRegexpMeta)
					assert.Equal(t, value == "true", *cfg.Server.FTP.EscapeRegexpMeta)
					assert.Contains(t, output.String(), "was removed and is ignored")
				}
			})
		}
	}
}
func TestLoadSFTPKnownHostsFile(t *testing.T) {
	for _, source := range []string{"file", "env", "missing"} {
		t.Run(source, func(t *testing.T) {
			knownHostsFile := filepath.Join(t.TempDir(), "known_hosts")
			if source != "missing" {
				require.NoError(t, os.WriteFile(knownHostsFile, nil, 0o600))
			}
			filename := "./fixtures/config.sftp.yml"
			if source == "file" {
				content, err := os.ReadFile(filename)
				require.NoError(t, err)
				// Quoting with %q also escapes Windows path separators for YAML.
				content = bytes.Replace(content, []byte("  sftp:"), []byte(fmt.Sprintf("  sftp:\n    knownHostsFile: %q", knownHostsFile)), 1)
				filename = filepath.Join(t.TempDir(), "ftpgrab.yml")
				// #nosec G703 -- The path is a test-owned temporary directory plus a fixed filename.
				require.NoError(t, os.WriteFile(filename, content, 0o600))
			} else {
				t.Setenv("FTPGRAB_SERVER_SFTP_KNOWNHOSTSFILE", knownHostsFile)
			}
			cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
			if source == "missing" {
				require.ErrorContains(t, err, "KnownHostsFile")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, knownHostsFile, cfg.Server.SFTP.KnownHostsFile)
		})
	}
}

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
					Concurrency:   1,
					HideSkipped:   new(false),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
				Notif: &Notif{
					Apprise: &NotifApprise{
						Endpoint: "http://apprise:8000",
						Token:    "ftpgrab",
						Timeout:  new(10 * time.Second),
					},
					Discord: &NotifDiscord{
						WebhookURL:   "https://discord.com/api/webhooks/1234567890/token",
						RenderEmbeds: new(true),
						RenderFields: new(true),
						Timeout:      new(10 * time.Second),
					},
					Gotify: &NotifGotify{
						Endpoint: "https://gotify.example.com",
						Token:    "gotify-token",
						Priority: 1,
						Timeout:  new(10 * time.Second),
					},
					Mail: &NotifMail{
						Host:               "localhost",
						Port:               25,
						SSL:                new(false),
						InsecureSkipVerify: new(false),
						From:               "ftpgrab@example.com",
						To:                 []string{"webmaster@example.com"},
					},
					Ntfy: &NotifNtfy{
						Endpoint: "https://ntfy.sh",
						Topic:    "ftpgrab-reports",
						Priority: 3,
						Tags:     []string{"package"},
						Timeout:  new(10 * time.Second),
					},
					Script: &NotifScript{
						Cmd: "uname",
						Args: []string{
							"-a",
						},
					},
					Slack: &NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
						Timeout:    new(10 * time.Second),
					},
					Webhook: &NotifWebhook{
						Endpoint: "http://webhook.foo.com/sd54qad89azd5a",
						Proxy:    "http://proxy.example.com:3128",
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
	t.Cleanup(UnsetEnv("FTPGRAB_"))

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
					Concurrency:   1,
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
					Concurrency:   1,
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
	t.Cleanup(UnsetEnv("FTPGRAB_"))

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
					Concurrency:   1,
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
						To:                 []string{"webmaster@foo.com"},
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
					Concurrency:   1,
					HideSkipped:   new(true),
					TempFirst:     new(false),
					CreateBaseDir: new(false),
				},
				Notif: &Notif{
					Slack: &NotifSlack{
						WebhookURL: "https://hooks.slack.com/services/ABCD12EFG/HIJK34LMN/01234567890abcdefghij",
						Timeout:    new(10 * time.Second),
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

func TestSlackDeliveryConfig(t *testing.T) {
	tests := []struct {
		name      string
		timeout   string
		want      time.Duration
		wantError bool
	}{
		{
			name: "default timeout",
			want: 10 * time.Second,
		},
		{
			name:    "configured timeout",
			timeout: "2s",
			want:    2 * time.Second,
		},
		{
			name:      "zero timeout",
			timeout:   "0s",
			wantError: true,
		},
		{
			name:      "negative timeout",
			timeout:   "-1s",
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(UnsetEnv("FTPGRAB_"))
			t.Setenv("FTPGRAB_NOTIF_SLACK_WEBHOOKURL", "https://example.com/slack")
			t.Setenv("FTPGRAB_NOTIF_SLACK_PROXY", "http://localhost:3128")
			if tt.timeout != "" {
				t.Setenv("FTPGRAB_NOTIF_SLACK_TIMEOUT", tt.timeout)
			}
			cfg, err := Load(Cli{Cfgfile: "./fixtures/config.ftp.yml"}, Meta{})
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, *cfg.Notif.Slack.Timeout)
			assert.Equal(t, "http://localhost:3128", cfg.Notif.Slack.Proxy)
		})
	}
}

func TestWebhookProxyConfig(t *testing.T) {
	t.Cleanup(UnsetEnv("FTPGRAB_"))
	t.Setenv("FTPGRAB_NOTIF_WEBHOOK_PROXY", "http://localhost:3128")
	cfg, err := Load(Cli{Cfgfile: "./fixtures/config.test.yml"}, Meta{})
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:3128", cfg.Notif.Webhook.Proxy)
}

func TestDiscordConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "webhook_url")
	require.NoError(t, os.WriteFile(filename, []byte("https://example.com/discord"), 0o600))
	tests := []struct {
		name      string
		environ   map[string]string
		wantError string
	}{
		{
			name: "defaults",
			environ: map[string]string{
				"WEBHOOKURL": "https://example.com/discord",
			},
		},
		{
			name: "custom options",
			environ: map[string]string{
				"WEBHOOKURL":   "https://example.com/discord",
				"MENTIONS":     "<@123>,<@&456>",
				"RENDEREMBEDS": "false",
				"RENDERFIELDS": "false",
				"TIMEOUT":      "2s",
				"PROXY":        "http://localhost:3128",
			},
		},
		{
			name: "secret file",
			environ: map[string]string{
				"WEBHOOKURLFILE": filename,
			},
		},
		{
			name:      "missing URL",
			wantError: "WebhookURL",
		},
		{
			name: "missing secret file",
			environ: map[string]string{
				"WEBHOOKURLFILE": filename + ".missing",
			},
			wantError: "WebhookURLFile",
		},
		{
			name: "zero timeout",
			environ: map[string]string{
				"WEBHOOKURL": "https://example.com/discord",
				"TIMEOUT":    "0s",
			},
			wantError: "Timeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(UnsetEnv("FTPGRAB_"))
			t.Setenv("FTPGRAB_NOTIF_DISCORD_RENDEREMBEDS", "true")
			for key, value := range tt.environ {
				t.Setenv("FTPGRAB_NOTIF_DISCORD_"+key, value)
			}
			cfg, err := Load(Cli{Cfgfile: "./fixtures/config.ftp.yml"}, Meta{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg.Notif.Discord)
			discord := cfg.Notif.Discord
			assert.Equal(t, tt.environ["WEBHOOKURL"], discord.WebhookURL)
			assert.Equal(t, tt.environ["WEBHOOKURLFILE"], discord.WebhookURLFile)
			assert.Equal(t, tt.environ["PROXY"], discord.Proxy)
			if tt.name == "custom options" {
				assert.Equal(t, []string{"<@123>", "<@&456>"}, discord.Mentions)
				assert.False(t, *discord.RenderEmbeds)
				assert.False(t, *discord.RenderFields)
				assert.Equal(t, 2*time.Second, *discord.Timeout)
			} else {
				assert.True(t, *discord.RenderEmbeds)
				assert.True(t, *discord.RenderFields)
				assert.Equal(t, 10*time.Second, *discord.Timeout)
			}
		})
	}
}

func TestNtfyConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("secret-token"), 0o600))
	tests := []struct {
		name      string
		environ   map[string]string
		wantError string
	}{
		{
			name: "custom options",
			environ: map[string]string{
				"ENDPOINT":  "https://ntfy.example.com",
				"TOKEN":     "inline-token",
				"TOKENFILE": filename,
				"TOPIC":     "custom-reports",
				"PRIORITY":  "5",
				"TAGS":      "warning,package",
				"ICON":      "https://example.com/icon.png",
				"CLICK":     "https://example.com/reports",
				"TIMEOUT":   "2s",
				"PROXY":     "http://localhost:3128",
			},
		},
		{
			name: "missing topic",
			environ: map[string]string{
				"TOPIC": "",
			},
			wantError: "Topic",
		},
		{
			name: "invalid priority",
			environ: map[string]string{
				"PRIORITY": "6",
			},
			wantError: "Priority",
		},
		{
			name: "zero timeout",
			environ: map[string]string{
				"TIMEOUT": "0s",
			},
			wantError: "Timeout",
		},
		{
			name: "missing token file",
			environ: map[string]string{
				"TOKENFILE": filename + ".missing",
			},
			wantError: "TokenFile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(UnsetEnv("FTPGRAB_"))
			t.Setenv("FTPGRAB_NOTIF_NTFY_TOPIC", "reports")
			for key, value := range tt.environ {
				t.Setenv("FTPGRAB_NOTIF_NTFY_"+key, value)
			}
			cfg, err := Load(Cli{Cfgfile: "./fixtures/config.ftp.yml"}, Meta{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, &NotifNtfy{
				Endpoint:  "https://ntfy.example.com",
				Token:     "inline-token",
				TokenFile: filename,
				Topic:     "custom-reports",
				Priority:  5,
				Tags:      []string{"warning", "package"},
				Icon:      "https://example.com/icon.png",
				Click:     "https://example.com/reports",
				Timeout:   new(2 * time.Second),
				Proxy:     "http://localhost:3128",
			}, cfg.Notif.Ntfy)
		})
	}
}

func TestGotifyConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("gotify-token"), 0o600))
	tests := []struct {
		name      string
		environ   map[string]string
		wantError string
	}{
		{
			name: "defaults",
		},
		{
			name: "custom options",
			environ: map[string]string{
				"PRIORITY": "5",
				"TIMEOUT":  "2s",
				"PROXY":    "http://localhost:3128",
			},
		},
		{
			name: "secret file",
			environ: map[string]string{
				"TOKEN":     "",
				"TOKENFILE": filename,
			},
		},
		{
			name: "missing endpoint",
			environ: map[string]string{
				"ENDPOINT": "",
			},
			wantError: "Endpoint",
		},
		{
			name: "missing token",
			environ: map[string]string{
				"TOKEN": "",
			},
			wantError: "Token",
		},
		{
			name: "negative priority",
			environ: map[string]string{
				"PRIORITY": "-1",
			},
			wantError: "Priority",
		},
		{
			name: "zero timeout",
			environ: map[string]string{
				"TIMEOUT": "0s",
			},
			wantError: "Timeout",
		},
		{
			name: "missing secret file",
			environ: map[string]string{
				"TOKENFILE": filename + ".missing",
			},
			wantError: "TokenFile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(UnsetEnv("FTPGRAB_"))
			t.Setenv("FTPGRAB_NOTIF_GOTIFY_ENDPOINT", "https://gotify.example.com")
			t.Setenv("FTPGRAB_NOTIF_GOTIFY_TOKEN", "gotify-token")
			for key, value := range tt.environ {
				t.Setenv("FTPGRAB_NOTIF_GOTIFY_"+key, value)
			}
			cfg, err := Load(Cli{Cfgfile: "./fixtures/config.ftp.yml"}, Meta{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			gotify := cfg.Notif.Gotify
			assert.Equal(t, "https://gotify.example.com", gotify.Endpoint)
			if tt.name == "secret file" {
				assert.Empty(t, gotify.Token)
				assert.Equal(t, filename, gotify.TokenFile)
			} else {
				assert.Equal(t, "gotify-token", gotify.Token)
			}
			if tt.name == "custom options" {
				assert.Equal(t, 5, gotify.Priority)
				assert.Equal(t, 2*time.Second, *gotify.Timeout)
				assert.Equal(t, "http://localhost:3128", gotify.Proxy)
			} else {
				assert.Equal(t, 1, gotify.Priority)
				assert.Equal(t, 10*time.Second, *gotify.Timeout)
			}
		})
	}
}

func TestAppriseConfig(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(filename, []byte("ftpgrab-key"), 0o600))
	tests := []struct {
		name      string
		environ   map[string]string
		wantError string
	}{
		{
			name: "defaults",
		},
		{
			name: "custom options",
			environ: map[string]string{ //nolint:gosec // The Apprise config key is a test fixture, not a credential.
				"TOKEN":     "ftpgrab-key",
				"TOKENFILE": filename,
				"TAGS":      "reports,ops",
				"URLS":      "mailto://ops@example.com,ntfy://ntfy.sh/reports",
				"TIMEOUT":   "2s",
				"PROXY":     "http://localhost:3128",
			},
		},
		{
			name: "secret file",
			environ: map[string]string{
				"TOKENFILE": filename,
			},
		},
		{
			name: "missing endpoint",
			environ: map[string]string{
				"ENDPOINT": "",
			},
			wantError: "Endpoint",
		},
		{
			name: "zero timeout",
			environ: map[string]string{
				"TIMEOUT": "0s",
			},
			wantError: "Timeout",
		},
		{
			name: "missing secret file",
			environ: map[string]string{
				"TOKENFILE": filename + ".missing",
			},
			wantError: "TokenFile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(UnsetEnv("FTPGRAB_"))
			t.Setenv("FTPGRAB_NOTIF_APPRISE_ENDPOINT", "http://apprise:8000")
			for key, value := range tt.environ {
				t.Setenv("FTPGRAB_NOTIF_APPRISE_"+key, value)
			}
			cfg, err := Load(Cli{Cfgfile: "./fixtures/config.ftp.yml"}, Meta{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			apprise := cfg.Notif.Apprise
			assert.Equal(t, "http://apprise:8000", apprise.Endpoint)
			assert.Equal(t, tt.environ["TOKEN"], apprise.Token)
			assert.Equal(t, tt.environ["TOKENFILE"], apprise.TokenFile)
			assert.Equal(t, tt.environ["PROXY"], apprise.Proxy)
			if tt.name == "custom options" {
				assert.Equal(t, []string{"reports", "ops"}, apprise.Tags)
				assert.Equal(t, []string{"mailto://ops@example.com", "ntfy://ntfy.sh/reports"}, apprise.URLs)
				assert.Equal(t, 2*time.Second, *apprise.Timeout)
			} else {
				assert.Equal(t, 10*time.Second, *apprise.Timeout)
			}
		})
	}
}

func TestHTTPNotifierTLSConfig(t *testing.T) {
	files := []string{
		filepath.Join(t.TempDir(), "first.pem"),
		filepath.Join(t.TempDir(), "second.pem"),
	}
	for _, filename := range files {
		require.NoError(t, os.WriteFile(filename, nil, 0o600))
	}
	for _, provider := range []string{"apprise", "discord", "gotify", "ntfy", "slack", "webhook"} {
		for _, source := range []string{"file", "env", "missing"} {
			t.Run(provider+"/"+source, func(t *testing.T) {
				t.Cleanup(UnsetEnv("FTPGRAB_"))
				filename := "./fixtures/config.test.yml"
				if source == "file" {
					content, err := os.ReadFile(filename)
					require.NoError(t, err)
					anchor := "  " + provider + ":"
					replacement := fmt.Sprintf("%s\n    tlsSkipVerify: true\n    tlsCaCertFiles:\n      - %q\n      - %q", anchor, files[0], files[1])
					content = bytes.Replace(content, []byte(anchor), []byte(replacement), 1)
					filename = filepath.Join(t.TempDir(), "ftpgrab.yml")
					// #nosec G703 -- The file is a test-owned temporary configuration.
					require.NoError(t, os.WriteFile(filename, content, 0o600))
				} else {
					prefix := "FTPGRAB_NOTIF_" + strings.ToUpper(provider)
					t.Setenv(prefix+"_TLSSKIPVERIFY", "true")
					value := strings.Join(files, ",")
					if source == "missing" {
						value += ".missing"
					}
					t.Setenv(prefix+"_TLSCACERTFILES", value)
				}
				cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
				if source == "missing" {
					require.ErrorContains(t, err, "TLSCACertFiles")
					return
				}
				require.NoError(t, err)
				var skipVerify bool
				var caFiles []string
				switch provider {
				case "apprise":
					skipVerify, caFiles = cfg.Notif.Apprise.TLSSkipVerify, cfg.Notif.Apprise.TLSCACertFiles
				case "discord":
					skipVerify, caFiles = cfg.Notif.Discord.TLSSkipVerify, cfg.Notif.Discord.TLSCACertFiles
				case "gotify":
					skipVerify, caFiles = cfg.Notif.Gotify.TLSSkipVerify, cfg.Notif.Gotify.TLSCACertFiles
				case "ntfy":
					skipVerify, caFiles = cfg.Notif.Ntfy.TLSSkipVerify, cfg.Notif.Ntfy.TLSCACertFiles
				case "slack":
					skipVerify, caFiles = cfg.Notif.Slack.TLSSkipVerify, cfg.Notif.Slack.TLSCACertFiles
				case "webhook":
					skipVerify, caFiles = cfg.Notif.Webhook.TLSSkipVerify, cfg.Notif.Webhook.TLSCACertFiles
				}
				assert.True(t, skipVerify)
				assert.Equal(t, files, caFiles)
			})
		}
	}
}

func TestMailRecipients(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		env       string
		want      []string
		wantError bool
	}{
		{
			name: "single address",
			yaml: "ops@example.com",
			env:  "ops@example.com",
			want: []string{"ops@example.com"},
		},
		{
			name: "multiple addresses",
			yaml: "[ops@example.com, dev@example.com]",
			env:  "ops@example.com, dev@example.com",
			want: []string{"ops@example.com", "dev@example.com"},
		},
		{
			name:      "empty recipients",
			yaml:      "[]",
			wantError: true,
		},
		{
			name:      "invalid second address",
			yaml:      "[ops@example.com, broken]",
			env:       "ops@example.com,broken",
			wantError: true,
		},
		{
			name:      "empty second address",
			yaml:      "[ops@example.com, '']",
			env:       "ops@example.com,",
			wantError: true,
		},
	}
	for _, tt := range tests {
		for _, source := range []string{"file", "env"} {
			t.Run(tt.name+"/"+source, func(t *testing.T) {
				t.Cleanup(UnsetEnv("FTPGRAB_"))
				content, err := os.ReadFile("./fixtures/config.ftp.yml")
				require.NoError(t, err)
				content = append(content, []byte("\nnotif:\n  mail:\n    from: ftpgrab@example.com\n    localName: file.example.com\n")...)
				localName := "file.example.com"
				if source == "file" {
					content = append(content, []byte("    to: "+tt.yaml+"\n")...)
				} else {
					t.Setenv("FTPGRAB_NOTIF_MAIL_TO", tt.env)
					t.Setenv("FTPGRAB_NOTIF_MAIL_LOCALNAME", "env.example.com")
					localName = "env.example.com"
				}
				filename := filepath.Join(t.TempDir(), "ftpgrab.yml")
				// #nosec G703 -- The file is a test-owned temporary configuration.
				require.NoError(t, os.WriteFile(filename, content, 0o600))
				cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
				if tt.wantError {
					require.ErrorContains(t, err, "To")
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.want, cfg.Notif.Mail.To)
				assert.Equal(t, localName, cfg.Notif.Mail.LocalName)
			})
		}
	}
}

func TestMailRecipientsEnvOverride(t *testing.T) {
	t.Cleanup(UnsetEnv("FTPGRAB_"))
	t.Setenv("FTPGRAB_NOTIF_MAIL_TO", "ops@example.com,dev@example.com")
	cfg, err := Load(Cli{Cfgfile: "./fixtures/config.test.yml"}, Meta{})
	require.NoError(t, err)
	assert.Equal(t, []string{"ops@example.com", "dev@example.com"}, cfg.Notif.Mail.To)
}

func TestSlackWebhookURLFile(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "webhook_url")
	require.NoError(t, os.WriteFile(secretFile, []byte("https://example.com/slack"), 0o600))
	tests := []struct {
		name        string
		webhookURL  string
		webhookFile string
		wantError   string
	}{
		{
			name:       "inline URL",
			webhookURL: "https://example.com/slack",
		},
		{
			name:        "file URL",
			webhookFile: secretFile,
		},
		{
			name:        "both configured",
			webhookURL:  "https://example.com/slack",
			webhookFile: secretFile,
		},
		{
			name:      "neither configured",
			wantError: "WebhookURL",
		},
		{
			name:        "missing file",
			webhookFile: secretFile + ".missing",
			wantError:   "WebhookURLFile",
		},
	}
	for _, tt := range tests {
		for _, source := range []string{"file", "env"} {
			t.Run(tt.name+"/"+source, func(t *testing.T) {
				t.Cleanup(UnsetEnv("FTPGRAB_"))
				content, err := os.ReadFile("./fixtures/config.ftp.yml")
				require.NoError(t, err)
				content = append(content, []byte("\nnotif:\n  slack:\n    timeout: 10s\n")...)
				if source == "file" {
					content = append(content, []byte(fmt.Sprintf("    webhookURL: %q\n    webhookURLFile: %q\n", tt.webhookURL, tt.webhookFile))...)
				} else {
					t.Setenv("FTPGRAB_NOTIF_SLACK_WEBHOOKURL", tt.webhookURL)
					t.Setenv("FTPGRAB_NOTIF_SLACK_WEBHOOKURLFILE", tt.webhookFile)
				}
				filename := filepath.Join(t.TempDir(), "ftpgrab.yml")
				// #nosec G703 -- The file is a test-owned temporary configuration.
				require.NoError(t, os.WriteFile(filename, content, 0o600))
				cfg, err := Load(Cli{Cfgfile: filename}, Meta{})
				if tt.wantError != "" {
					require.ErrorContains(t, err, tt.wantError)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.webhookURL, cfg.Notif.Slack.WebhookURL)
				assert.Equal(t, tt.webhookFile, cfg.Notif.Slack.WebhookURLFile)
			})
		}
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
