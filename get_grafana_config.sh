#!/bin/bash -ex
ssh -i ~/.ssh/vultr root@build.danstutzman.com "sqlite3 /root/grafana/data/grafana.db 'select data from dashboard;'" > grafana_config.txt
