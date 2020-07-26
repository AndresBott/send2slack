# -*- mode: ruby -*-
# vi: set ft=ruby :


Vagrant.configure("2") do |config|

  config.vm.box = "debian/contrib-buster64"

  config.vm.synced_folder "./", "/home/vagrant/app"
  config.vm.provider "virtualbox" do |v|
    # set the name of the VM
    v.name = "send2Slack-dev"

    # use a linked clone of the imported machine
    v.linked_clone = true

    # use VBoxManage to make vm setting
    #v.customize ["modifyvm", :id, "--cpuexecutioncap", "50"]
    v.customize ["modifyvm", :id, "--ioapic", "on"]
    v.memory = 1024
    v.cpus = 1
  end

  #===========================================================================================================
  # Provision
  #===========================================================================================================

  config.vm.provision "shell", path: "vagrant/vagrantSetup.bash"


end
