#!/usr/bin/env bash
read -sp "Enter your git message : " message
go build -o filter
git add .
git commit -m "$message"
git push
