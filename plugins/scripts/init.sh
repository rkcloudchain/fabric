#!/bin/bash
#
# Copyright Greg Haskins All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Update the entire system to the latest releases

echo "deb http://mirrors.tencentyun.com/ubuntu/ xenial main restricted universe multiverse" > /etc/apt/sources.list
echo "deb http://mirrors.tencentyun.com/ubuntu/ xenial-security main restricted universe multiverse" >> /etc/apt/sources.list
echo "deb http://mirrors.tencentyun.com/ubuntu/ xenial-updates main restricted universe multiverse" >> /etc/apt/sources.list
echo "deb-src http://mirrors.tencentyun.com/ubuntu/ xenial main restricted universe multiverse" >> /etc/apt/sources.list
echo "deb-src http://mirrors.tencentyun.com/ubuntu/ xenial-security main restricted universe multiverse" >> /etc/apt/sources.list
echo "deb-src http://mirrors.tencentyun.com/ubuntu/ xenial-updates main restricted universe multiverse" >> /etc/apt/sources.list

apt-get update
apt-get upgrade -y
apt-get dist-upgrade -y
apt-get install -y wget tzdata