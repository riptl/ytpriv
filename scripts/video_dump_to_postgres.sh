#!/bin/bash
jq -r '"INSERT INTO videos (id, uploader_id, duration, genre) VALUES ($$\(.id)$$, $$\(.uploader_id)$$, \(.duration), $$\(.genre)$$) ON CONFLICT DO NOTHING;"' $@
