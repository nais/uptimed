#!/usr/bin/env bash
COMMIT_HASH=$1

curl --user ${CIRCLECI_TOKEN}: \
    --request POST \
    --form revision=${COMMIT_HASH}\
    --form config=@config.yml \
    --form notify=false \
        https://circleci.com/api/v1.1/project/github/nais/uptimed/tree/master
