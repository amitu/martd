import urllib.request
import argparse
import uuid
import json


parser = argparse.ArgumentParser()
parser.add_argument("channel")
parser.add_argument("--endpoint", default="localhost:54321")
parser.add_argument("--etag", default="0")
args = parser.parse_args()

cid = str(uuid.uuid4())
etag = args.etag

while True:
    try:
        d = json.loads(
            urllib.request.urlopen(
                "http://%s/sub?cid=%s&%s=%s" % (
                    args.endpoint, cid, args.channel, etag
                )
            ).read().decode("utf-8")
        )
    except Exception as e:
        print(e)
        break
    except KeyboardInterrupt:
        print("Ctrl-C! Sure.")
        break
    etag = d["channels"][args.channel]["etag"]
    for payload in d["channels"][args.channel]["payload"]:
        print(payload)
