# Telegraf Output Plugin for Amazon Kinesis Data Firehose

[![](https://img.shields.io/github/license/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](LICENSE)
[![](https://img.shields.io/github/actions/workflow/status/muhlba91/telegraf-output-kinesis-data-firehose/verify.yml?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/actions/workflows/verify.yml)
[![](https://img.shields.io/coveralls/github/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/)
[![](https://img.shields.io/github/release-date/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/releases)
[![](https://img.shields.io/github/downloads/muhlba91/telegraf-output-kinesis-data-firehose/total?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/releases)
<a href="https://www.buymeacoffee.com/muhlba91" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="28" width="150"></a>

This plugin makes use of the [Telegraf Output Exec plugin](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/exec).
It will batch up Points in one Put request to Amazon Kinesis Data Firehose.

The plugin also provides optional common formatting options, like normalizing keys and flattening the output.
Such configuration can be used to provide data ingestion without the need of a data transformation function.

It expects that the configuration for the output ship data in line format.

---

## About Amazon Kinesis Data Firehose

It may be useful for users to review Amazons official documentation which is
available [here](https://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html).

## Usage

- Download the [latest release package](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/releases/latest) for your platform.

- Unpack the build to your system:

```bash
mkdir /var/lib/telegraf/firehose
chown telegraf:telegraf /var/lib/telegraf/firehose
tar xf telegraf-output-kinesis-data-firehose-<LATEST_VERSION>-<OS>-<ARCH>.tar.gz -C /var/lib/telegraf/firehose
# e.g. tar xf telegraf-output-kinesis-data-firehose-v1.0.0-linux-amd64.tar.gz -C /var/lib/telegraf/firehose
```

- Edit the plugin configuration as needed:

```bash
vi /var/lib/telegraf/firehose/plugin.conf
```

- Add the plugin to `/etc/telegraf/telegraf.conf` or into a new file in `/etc/telegraf/telegraf.d`:

```toml
[[outputs.exec]]
  command = [ "/var/lib/telegraf/firehose/telegraf-output-kinesis-data-firehose", "-config", "/var/lib/telegraf/firehose/plugin.conf" ]
  data_format = "influx"
```

- Restart or reload Telegraf.

## AWS Authentication

This plugin uses a credential chain for Authentication with the Amazon Kinesis Data Firehose API
endpoint. The plugin will attempt to authenticate in the following order:

1. web identity provider credentials via STS if `role_arn` and
   `web_identity_token_file` are specified,
2. assumed credentials via STS if the `role_arn` attribute is specified (source
   credentials are evaluated from subsequent rules),
3. explicit credentials from the `access_key`, and `secret_key` attributes,
4. shared profile from the `profile` attribute,
5. environment variables,
6. shared credentials, and/or
7. the EC2 instance profile.

If you are using credentials from a web identity provider, you can specify the
session name using `role_session_name`.
If left empty, the current timestamp will be used.

## Configuration

```toml @plugin.conf
## AWS region
region = "eu-west-1"

## AWS credentials
#access_key = ""
#secret_key = ""
#role_arn = ""
#web_identity_token_file = ""
#role_session_name = ""
#profile = ""
#shared_credential_file = ""

## Endpoint to make request against, the correct endpoint is automatically
## determined and this option should only be set if you wish to override the
## default.
##   ex: endpoint_url = "http://localhost:8000"
#endpoint_url = ""

## Amazon Kinesis Data Firehose DeliveryStreamName must exist prior to starting telegraf.
streamname = "DeliveryStreamName"

## 'debug' will show upstream AWS messages.
debug = false

## 'format' provides formatting options
#[format]
  ## 'flatten' flattens all tags and fields into top-level keys
  #flatten = false
  ## 'normalize_keys' normalizes all keys to:
  ## 1/ convert to lower case, and
  ## 2/ replace spaces (' ') with underscores ('_')
  #normalize_keys = false
  ## 'name_key_rename' renames the 'name' field to the provided value
  #name_key_rename = ""
  ## 'timestamp_as_rfc3339' parses the timestamp into RFC3339 instead of a unix timestamp
  #timestamp_as_rfc3339 = false
  ## 'timestamp_units' defines the unix timestamp precision
  #timestamp_units = "1ms"
```

For this output plugin to function correctly the following variables must be configured:

- region: the AWS region to connect to
- streamname: used to send data to the correct stream (the stream *MUST* be pre-configured prior to starting this plugin!)

---

## Development

The project uses go modules which can be downloaded by running:

```bash
go mod download
```

### Testing

1) Install all dependencies as shown above.
2) Run `go test` by:

```bash
make test
```

### Linting and Code Style

The project uses [golangci-lint](http://golangci-lint.run), and also [pre-commit](https://pre-commit.com/).

1) Install all dependencies as shown above.
2) (Optional) Install pre-commit hooks:

```bash
pre-commit install
```

3) Run linter:

```bash
make lint
```

### Commit Message

This project follows [Conventional Commits](https://www.conventionalcommits.org/), and your commit message must also
adhere to the additional rules outlined in `.conform.yaml`.

---

## Contributions

Please feel free to contribute, be it with Issues or Pull Requests! Please read
the [Contribution guidelines](CONTRIBUTING.md)

## Notes

The plugin was inspired by the [Amazon Kinesis Data Stream Output Plugin](https://github.com/morfien101/telegraf-output-kinesis).

## Supporting

If you enjoy the application and want to support my efforts, please feel free to buy me a coffe. :)

<a href="https://www.buymeacoffee.com/muhlba91" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="75" width="300"></a>
