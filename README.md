# prstats

this is an AWS serverless app that records PR stats into Cloudwatch Metrics.

## build
./buildlambda.sh

## test
AWS API keys are needed in order to push metrics to cloudwatch.

    aws-vault exec <myaccount> sam local start-api
		ngrok http 3000

Then go to Github and configure a webhook to send the PR event, and configure the secret to match what you have set in template.yaml.

## TODO

* figure out sam package & deploy
* add appropriate Role/Policy for Lambda to execute with
* add Cloudwatch Dashboard into Template

