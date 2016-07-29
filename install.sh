#!/bin/bash
if [ -e "VERSION" ]
then
	zhmVersion=`cat VERSION`
else
	zhmVersion=devel
fi

zhmRevision=`git log -1 --pretty=format:"%h(%ai"`
zhmRevision="${zhmRevision:0:18})"

go install -ldflags "-X main.Version=$zhmVersion -X main.Revision=$zhmRevision"
