import urllib2
import httplib
import argparse
import uuid
import json
import time

parser = argparse.ArgumentParser()
parser.add_argument("channel")
parser.add_argument("--endpoint", default="localhost:54321")
args = parser.parse_args()

cid = str(uuid.uuid4())
etag = "0"

while True:
    try:
        d = json.loads(
            urllib2.urlopen(
                "http://%s/sub?cid=%s&%s=%s" % (
                    args.endpoint, cid, args.channel, etag
                )
            ).read()
        )
    except httplib.BadStatusLine, e:
        print e
        continue
    except urllib2.URLError, e:
        print e
        time.sleep(1)
        continue
    except KeyboardInterrupt:
        print "Ctrl-C! Sure."
        break
    etag = d["channels"][args.channel]["etag"]
    for payload in d["channels"][args.channel]["payload"]:
        print payload
