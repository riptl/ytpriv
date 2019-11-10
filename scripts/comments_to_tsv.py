#!/usr/bin/env python3

import fileinput
import json
import sys

#print("id\tvideo_id\tauthor_id\tcrawled_at\tcreated_before\tcreated_after\tlikes\tauthor\tcontent")
print("id\tvideo_id\tauthor_id\tcrawled_at\tlikes\tauthor\tcontent")


# noinspection SpellCheckingInspection
def process_line(l):
    obj = json.loads(l)

    oid = obj["id"]
    vid = obj["video_id"]
    aid = obj["author_id"]
    cwat = obj["crawled_at"]
    #crbf = obj["created_before"]
    #craf = obj["created_after"]
    likes = obj["likes"]
    aname = obj["author"]

    cobj = obj["content"]
    for cem in cobj:
        if "navigationEndpoint" in cem:
            nav = cem["navigationEndpoint"]
            del nav["clickTrackingParams"]
            del nav["commandMetadata"]
    content = json.dumps(cobj)

    print(f"{oid}\t{vid}\t{aid}\t{cwat}\t{likes}\t{aname}\t{content}")


i = 0
for line in fileinput.input():
    i += 1
    try:
        process_line(line)
    except Exception as e:
        print(f"Failed to parse line {i}: {e}", file=sys.stderr)
