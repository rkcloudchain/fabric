#!/bin/bash
#
# Copyright Greg Haskins All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# Update the entire system to the latest releases

apt-get update
apt-get upgrade -y
apt-get dist-upgrade -y
apt-get install -y wget tzdata