#!/bin/bash
set -e

ln -s "$(readlink -f $(dirname $0)/backup.sh)" /etc/cron.weekly
