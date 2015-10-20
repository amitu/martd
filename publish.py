import urllib2
import argparse

parser = argparse.ArgumentParser()
parser.add_argument("channel")
parser.add_argument("message")
parser.add_argument("--endpoint", default="localhost:54321")
parser.add_argument("--size", default=10, type=int)
args = parser.parse_args()

print urllib2.urlopen(
    "http://%s/pub?channel=%s&size=%s" % (
        args.endpoint, args.channel, args.size
    ), args.message
).read()
