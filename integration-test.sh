#!/bin/bash

set -e

HELP='
usage: integration-test.sh [<flags>]

Runs integration tests that send realistic HTTP requests to API Gateway via
SAM local.

Flags:
  --help        Show help
  --localstack  Run Cloudwatch via localstack in docker and hook SDK
                to send metrics to localstack endpoint.

Dependencies:
  AWS Serverless Application Model (SAM)
  Docker
  Go (for building lambda function)
'

function help() {
  echo "$HELP" >&2
  exit 1
}

function missing() {
  echo >&2 "Error: $1 must be installed. Aborting."
  help
}

localstack=0

while :; do
	case $1 in 
		-h|-\?|--help)
			help
			;;
		-l|--local)
			localstack=1
			;;
		--) # end of all options
			shift
			break
			;;
		-?*)
			printf 'WARN: Unknown option (ignored): %s\n' "$1" >&2
			;;
		*) # no more options so break loop
			break
	esac
	shift
done

type sam >/dev/null 2>&1 || missing "sam cli"

echo 'Running sam local invoke against requests in test/*.json'

# TODO: report on test case failures

find 'test' -name '*.json' | xargs -n1 sam local invoke --event

