#!/usr/bin/env bash
#=============================================================================================================
## add a better bashrc for root shell
#=============================================================================================================
sudo cp /vagrant/vagrant/rootbashrc /root/.bashrc
sudo chown root:root /root/.bashrc

sudo cp /vagrant/vagrant/vagrantbashrc /home/vagrant/.bashrc
sudo chown vagrant:vagrant /home/vagrant/.bashrc

#=============================================================================================================
## install apps
#=============================================================================================================
sudo apt-get update
sudo apt-get install joe curl wget



cd /home/vagrant
rm -rf /usr/local/go/
wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
tar -xvf go1.13.3.linux-amd64.tar.gz
sudo mv go /usr/local

## goreleaser
wget https://github.com/goreleaser/goreleaser/releases/download/v0.131.1/goreleaser_amd64.deb
sudo dpkg -i goreleaser_amd64.deb


## some folders and data
sudo mkdir /etc/send2slack
sudo ln -s /home/vagrant/app/dist/send2slack_linux_amd64/send2slack /usr/local/sbin/sendmail


## add cron to send a message very minute
line="* * * * * echo test from cron || false"
(crontab -u vagrant -l; echo "$line" ) | crontab -u vagrant -
