#!/bin/bash
PATH=${HOME}/go/bin:${PATH}
godoc -http=:6060 -goroot /usr/share/go
