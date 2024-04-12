#!/bin/bash

for dev in $(ip link show | grep 'tap' | awk '{print $2}' | sed 's/://'); do sudo ip link delete $dev; done