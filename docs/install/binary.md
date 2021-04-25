# Installation from binary

## Download

FTPGrab binaries are available on [releases]({{ config.repo_url }}releases/latest) page.

Choose the archive matching the destination platform:

* [`ftpgrab_{{ git.tag | trim('v') }}_darwin_arm64.tar.gz`]({{ config.repo_url }}releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_darwin_arm64.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_darwin_x86_64.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_darwin_x86_64.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_freebsd_i386.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_freebsd_i386.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_freebsd_x86_64.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_freebsd_x86_64.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_arm64.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_arm64.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_armv5.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_armv5.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_armv6.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_armv6.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_armv7.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_armv7.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_ppc64le.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_ppc64le.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_s390x.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_s390x.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_i386.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_i386.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz)
* [`ftpgrab_{{ git.tag | trim('v') }}_windows_i386.zip`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_windows_i386.zip)
* [`ftpgrab_{{ git.tag | trim('v') }}_windows_x86_64.zip`]({{ config.repo_url }}/releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_windows_x86_64.zip)

And extract FTPGrab:

```shell
wget -qO- {{ config.repo_url }}releases/download/v{{ git.tag | trim('v') }}/ftpgrab_{{ git.tag | trim('v') }}_linux_x86_64.tar.gz | tar -zxvf - ftpgrab
```

After getting the binary, it can be tested with [`./ftpgrab --help`](../usage/cli.md) command and moved to a permanent
location.

## Server configuration

Steps below are the recommended server configuration.

### Prepare environment

Create user to run FTPGrab (ex. `ftpgrab`)

```shell
groupadd ftpgrab
useradd -s /bin/false -d /bin/null -g ftpgrab ftpgrab
```

### Create required directory structure

```shell
mkdir -p /var/lib/ftpgrab
chown ftpgrab:ftpgrab /var/lib/ftpgrab/
chmod -R 750 /var/lib/ftpgrab/
mkdir /etc/ftpgrab
chown ftpgrab:ftpgrab /etc/ftpgrab
chmod 770 /etc/ftpgrab
```

### Configuration

Create your first [configuration](../config/index.md) file in `/etc/ftpgrab/ftpgrab.yml` and type:

```shell
chown ftpgrab:ftpgrab /etc/ftpgrab/ftpgrab.yml
chmod 644 /etc/ftpgrab/ftpgrab.yml
```

### Copy binary to global location

```shell
cp ftpgrab /usr/local/bin/ftpgrab
```

## Running FTPGrab

After the above steps, two options to run FTPGrab:

### 1. Creating a service file (recommended)

See how to create [Linux service](linux-service.md) to start FTPGrab automatically.

### 2. Running from terminal

```shell
FTPGRAB_DB_PATH=/var/lib/ftpgrab/ftpgrab.db /usr/local/bin/ftpgrab \
  --config /etc/ftpgrab/ftpgrab.yml \
  --schedule "*/30 * * * *"
```

## Updating to a new version

You can update to a new version of FTPGrab by stopping it, replacing the binary at `/usr/local/bin/ftpgrab` and
restarting the instance.

If you have carried out the installation steps as described above, the binary should have the generic name `ftpgrab`.
Do not change this, i.e. to include the version number.
