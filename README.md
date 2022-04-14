# Kentik Custom Notification Templates

Kentik Portal allows its users to setup notification channels of different types, including email, Slack, Microsoft Teams and others. In order to further customize the notifications experience and integrate Kentik with various 3rd party platforms, users can take advantage of Custom Webhook notification channels.

This repository is meant to provide documentation and examples of custom notification templates that can be used to format the payload sent to Custom Webhook notification channels.

To understand what notifications are, how to use and configure them and how to initially setup custom webhook notification channels, please refer to the [Kentik Knowledge Base](https://kb.kentik.com/v4/Ca00.htm).

## Example Templates

Within the [Templates](templates/) directory you can find ready, baseline custom webhook templates that can be used right away or customized further according to your needs.

## Using Custom Webhook Templating

[Using Custom Webhook Templating](docs/TEMPLATING_REFERENCE.md) provides comprehensive details on how notifications are being rendered and what syntax, variables, methods and functions are available. It can help you get familiar with how to develop new templates and how to customize your existing ones.

## How to Develop and Test New Template

The repository also provides a means to test and validate that a given template will work well with the Kentik notifications engine. For more details please refer to a separate article on [How to Develop and Test New Templates](docs/DEVELOPERS_GUIDE.md).
