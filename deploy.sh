#!/bin/bash -ex

pushd $GOPATH/src/github.com/danielstutzman/prometheus-email-reports
go vet .
pushd

go install -race .

fwknop -s -n monitoring.danstutzman.com
ssh root@monitoring.danstutzman.com <<"EOF"
  set -ex

  id -u prometheus-email-reports &>/dev/null || sudo useradd prometheus-email-reports
  sudo mkdir -p /home/prometheus-email-reports
  sudo chown prometheus-email-reports:prometheus-email-reports /home/prometheus-email-reports
  cd /home/prometheus-email-reports

  GOROOT=/home/prometheus-email-reports/go1.7.3.linux-amd64
  if [ ! -e $GOROOT ]; then
    sudo curl https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz >go1.7.3.linux-amd64.tar.gz
    chown prometheus-email-reports:prometheus-email-reports go1.7.3.linux-amd64.tar.gz
    sudo -u prometheus-email-reports tar xzf go1.7.3.linux-amd64.tar.gz
    sudo -u prometheus-email-reports mv go $GOROOT
  fi
  GOPATH=/home/prometheus-email-reports/gopath
  sudo -u prometheus-email-reports mkdir -p $GOPATH
  sudo -u prometheus-email-reports mkdir -p $GOPATH/src/github.com/danielstutzman/prometheus-email-reports
EOF

time rsync -a -e "ssh -C" -r . root@monitoring.danstutzman.com:/home/prometheus-email-reports/gopath/src/github.com/danielstutzman/prometheus-email-reports --include='*.go' --include='*/' --exclude='*' --prune-empty-dirs --delete

fwknop -s -n monitoring.danstutzman.com
ssh root@monitoring.danstutzman.com <<"EOF"
  set -ex

  GOROOT=/home/prometheus-email-reports/go1.7.3.linux-amd64
  GOPATH=/home/prometheus-email-reports/gopath
  cd $GOPATH/src/github.com/danielstutzman/prometheus-email-reports
  chown -R prometheus-email-reports:prometheus-email-reports .
  time sudo -u prometheus-email-reports GOPATH=$GOPATH GOROOT=$GOROOT $GOROOT/bin/go install -race

  touch /var/log/prometheus-email-reports.log
  chown prometheus-email-reports:root /var/log/prometheus-email-reports.log
  tee /etc/cron.d/prometheus-email-reports <<EOF2
0 7 * * * prometheus-email-reports /home/prometheus-email-reports/gopath/bin/prometheus-email-reports -pngPath out.png -prometheusHostPort localhost:9090 -emailFrom "Reports <reports@monitoring.danstutzman.com>" -emailSubject "Report with Prometheus metrics" -emailTo dtstutz@gmail.com -smtpHostPort localhost:25 >> /var/log/prometheus-email-reports.log
EOF2
EOF

cat <<EOF
To test, run:
su -l prometheus-email-reports
cat /etc/cron.d/prometheus-email-reports # to see command to run
exit
EOF
