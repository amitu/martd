import urllib2
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("channel")
parser.add_argument("message")
parser.add_argument("--endpoint", default="localhost:54321")
parser.add_argument("--size", default=10, type=int)
parser.add_argument("--one2one", default=False, type=bool)
args = parser.parse_args()

print urllib2.urlopen(
    "http://%s/pub?channel=%s&size=%s&one2one=%s" % (
        args.endpoint, args.channel, args.size,
        "true" if args.one2one else "false"
    ), args.message
).read()
