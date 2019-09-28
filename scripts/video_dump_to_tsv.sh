#!/bin/bash
jq -r '"\(.id)\t\(.uploader_id)\t\(.duration)\t\(.genre)"' $@
