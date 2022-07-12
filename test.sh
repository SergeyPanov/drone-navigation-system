#!/bin/bash

cd /drone-navigation-system || exit 1
go build -o /opt/drone-navigation-system/dns

/opt/drone-navigation-system/dns
