#!/bin/bash
set -e

read -p "Username: " user
read -sp "Password: " password

curl https://$user:$password@api.bitbucket.org/1.0/user/repositories \
    | python -m json.tool > bitbucket_index.json

jq '.[]
    | select(.scm == "git")
    | "git@bitbucket.org:" + .owner + "/" + .slug + ".git"' bitbucket_index.json \
    | xargs -P0 -n1 git clone --mirror

# TODO: mercurial repos
