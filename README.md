# seahorse
Automated dosing system for hydroponic reservoirs

## Overview
This repo contains the control software for automatically reading sensor data and activating pumps.
Currently, it has only been built for and tested on the [Raspberry Pi Zero W](https://www.raspberrypi.org/products/raspberry-pi-zero-w/).

## Setup
Some extra setup needs to be done to get this working on the Pi Zero W:
```
# if wanting to use UFW (requires reboot)
sudo update-alternatives --set iptables /usr/sbin/iptables-legacy

# Interface Options -> I2C -> Enable
sudo raspi-config

# to help debug I2C devices
sudo apt install i2c-tools
```
