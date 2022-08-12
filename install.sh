#!/usr/bin/env bash

if [ $1 = "uninstall" ]; then
  sudo apt remove --purge -y jailsh
  exit
fi

wget https://github.com/CodeRadu/jail-shell/releases/download/0.0.1/jailsh_0.0.1_amd64.deb -O jailsh.deb
sudo apt install -y ./jailsh.deb
rm jailsh.deb
echo Done
