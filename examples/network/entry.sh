#!/bin/bash

/usr/bin/ros-wati-for -d --containers network --interfaces eth0

exec "$@"
