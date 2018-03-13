#!/bin/sh

/usr/bin/ros-wait-for -d --containers network --interfaces eth0

exec "$@"
