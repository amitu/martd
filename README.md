martd
=====


martd is a push server.






## Why not socketio/websocket?


Websocket requires server code. This is meant to be a pluggable solution.






## How it works?


martd could be hosted behind nginx, or can run as a standalone server. 
Supports both HTTP and HTTPS.

Browser, or other clients connect to martd and wait for events. A JS library
is available as part of this server, /client.js.

The emphasis is on supporting all possible clients/browsers.

Clients wait on data on "channels". Channels have data retention, and client
related attributes. Channels are created on from server side, IP whitelisting,
or shared cookie. Channels can be pushed into by server or a client with proper
key (one of the attributes of channel).

This library is meant for modest amount of data push from server to 
browser/client, it may not be the best choise for data heavey game like 
situations.

A demo will be available at http://martd.amitu.com






## Client API


handle = martd.sub("channel", function(){ /*subscribed*/ }, etag /*optional*/);
martd.unsub("channel", function(){ /*unsubscribed*/ });
handle = martd.on("ch", function(d){}); // .on() should be called before .sub()
martd.pub(channel, key, data)

handle.cancel() cancles subscription and data binding.

martd.client_id is a uniq id generated on each page load.







## Channels


Clients subscribe to "channels". Each client can connect to one or more 
channels.

Channels can have types. The first "push" to channel sets the channel 
attributes.

.size=1, max data in channel is stored in a "circular queue". Oldest messages
         are dropped to make way for new ones.

.life=600s, max life of data in channel.

.one2one, only one client allowed in this channel, subsequent clients are 
         rejected. If more than one are already connected when this attribute is
         being set, first one is left and rest ones are kicked out.

.key=key, unique key that acts like password for this channel, all push require
         this key.

Each push to channel must contain all attributes, as channel can be dropped 
anytime, whenever there is no data left in channel and no client is connected.






## Push


Push can be sent from "server side". An empty push can be sent to set channel
properties. This "empty" message would be kept in queue, but would not be sent
to connected clients.

Each push changes the etag for the channel. etag is sent to client to keep track
of seen status of a message.






## Proxy Pass


For production, this server should be configured behind nginx, on the same 
domain as main website.

For testing, this server takes a server location as command line argument, and
proxy passes everything but the http requests it is interested in.

For development, it comes packaged with a SSL certificate. If proxy pass feature
is being used with prod, you can pass your own SSL certificate.


## References

- https://github.com/wandenberg/nginx-push-stream-module/tree/master/docs/examples
- https://groups.google.com/forum/#!topic/golang-nuts/rY4KoouaQu4
- https://gist.github.com/nono/1048668
