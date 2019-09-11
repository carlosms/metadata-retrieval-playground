#! /bin/bash

go build -o metadata cmd/metadata/main.go

#export SOURCED_GITHUB_TOKEN=
export LOG_LEVEL=debug

mkdir bench

echo 'go-errors with v3'
for i in {1..5}; do ./metadata v3 --owner src-d --name go-errors > ./bench/go-errors.v3.stdout.$i.txt 2> ./bench/go-errors.v3.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done

echo 'sourced-ce with v3'
for i in {1..5}; do ./metadata v3 --owner src-d --name sourced-ce > ./bench/sourced-ce.v3.stdout.$i.txt 2> ./bench/sourced-ce.v3.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done

echo 'go-git with v3'
for i in {1..5}; do ./metadata v3 --owner src-d --name go-git > ./bench/go-git.v3.stdout.$i.txt 2> ./bench/go-git.v3.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done


echo 'go-errors with v4'
for i in {1..5}; do ./metadata v4 --owner src-d --name go-errors > ./bench/go-errors.v4.stdout.$i.txt 2> ./bench/go-errors.v4.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done

echo 'sourced-ce with v4'
for i in {1..5}; do ./metadata v4 --owner src-d --name sourced-ce > ./bench/sourced-ce.v4.stdout.$i.txt 2> ./bench/sourced-ce.v4.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done

echo 'go-git with v4'
for i in {1..5}; do ./metadata v4 --owner src-d --name go-git > ./bench/go-git.v4.stdout.$i.txt 2> ./bench/go-git.v4.stderr.$i.txt && break || echo 'command failed, retrying' && sleep 15; done

echo 'all done :)'
