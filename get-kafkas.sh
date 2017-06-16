#!/bin/bash
# uses knife and formats kafka addresses into flag for kcli

if [ "$#" -ne 3 ]; then
  >&2 echo "usage: chef-role chef-environment port"
  exit 1
fi

role=${1}
env=${2}
port=${3}

function ksn() {
	knife search node "roles:${1} AND chef_environment:${2}" -i
}

kafkas_w_commas=$(tr '\n' ',' < <(ksn ${role} ${env} | grep '.sendgrid.net' | grep 'kafka'))
kafkas_wo_last_comma=${kafkas_w_commas%?}

if [ ${#kafkas_wo_last_comma} -eq 0 ]; then
  >&2 echo "no kafkas found with chef-role: ${role}, chef-environment: ${env}"
  exit 1
fi

echo ${kafkas_wo_last_comma} | sed -e "s/.sendgrid.net/.sendgrid.net:${port}/g"
