# Telegraf Output Plugin for Amazon Kinesis Data Firehose

[![](https://img.shields.io/github/license/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](LICENSE)
[![](https://img.shields.io/github/workflow/status/muhlba91/telegraf-output-kinesis-data-firehose/Release?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/actions)
[![](https://img.shields.io/coveralls/github/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/)
[![](https://img.shields.io/github/release-date/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/releases)
[![](https://img.shields.io/github/downloads/muhlba91/telegraf-output-kinesis-data-firehose/total?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/releases)
[![](https://img.shields.io/snyk/vulnerabilities/github/muhlba91/telegraf-output-kinesis-data-firehose?style=for-the-badge)](https://github.com/muhlba91/telegraf-output-kinesis-data-firehose/)
<a href="https://www.buymeacoffee.com/muhlba91" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="28" width="150"></a>

This is a Telegraf output plugin that is still in the early stages of development.
This plugin makes use of the [Telegraf Output Exec plugin](https://github.com/influxdata/telegraf/tree/master/plugins/outputs/exec).
It will batch up Points in one Put request to Amazon Kinesis Data Firehose.

It expects that the configuration for the output ship data in line format. Here is an example configuration of the Exec Output.

```toml
[[outputs.exec]]
  command = [ "./telegraf_kinesis_output", "-config", "/path/to/plugin.conf" ]
  data_format = "influx"
```

## About Amazon Kinesis Data Firehose

It may be useful for users to review Amazons official documentation which is
available [here](https://docs.aws.amazon.com/firehose/latest/dev/what-is-this-service.html).

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
```

For this output plugin to function correctly the following variables must be configured:

* region
* streamname

### region

The region is the Amazon region that you wish to connect to.

### streamname

The streamname is used by the plugin to ensure that data is sent to the correct
Kinesis Data Firehose stream. It is important to note that the stream *MUST* be
pre-configured for this plugin to function correctly.
If the stream does not exist the plugin will result in
telegraf exiting with an exit code of 1.
