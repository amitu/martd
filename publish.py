import urllib.request
import argparse

SECOND = 1000 * 1000 * 1000  # nano seconds

parser = argparse.ArgumentParser()
parser.add_argument("channel")
parser.add_argument("message")
parser.add_argument("--endpoint", default="localhost:54321")
parser.add_argument("--size", default=10, type=int)
parser.add_argument("--life", default=60 * 60, type=int)
parser.add_argument("--one2one", default=False, action="store_true")
args = parser.parse_args()

print(urllib.request.urlopen(
    "http://%s/pub?channel=%s&size=%s&one2one=%s&life=%s" % (
        args.endpoint, args.channel, args.size,
        "true" if args.one2one else "false",
        args.life * SECOND
    ), args.message.encode("utf-8")
).read())
