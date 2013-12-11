# RProxy

rproxy was first invented by Andrew Tridgell, following his seminal work
on [rsync](https://rsync.samba.org/). rsync has become a central piece
of software for all unix-like users, allowing everyone to synchronise 2
computers while sending the minimum amount of data on the wire, thanks
to simple and smart techniques. If you haven't already, I suggest you
read [the paper](http://samba.org/~tridge/phd_thesis.pdf) describing how
rsync works; everyone can understand it, and everyone can learn from it.

[rproxy](https://rproxy.samba.org/) is an adaptation of the rsync
algorithm to the web, in order to reduce overall bandwidth consumption.

## How it works

(_For a more in-depth description, don't hesitate to read the original
 project's page_)

The idea is to have a pair of proxies between the client and the server.

One of the proxy is located close to the client (by close, I mean that
the network is virtually free and fast; think LAN) while the other
one is close to the server. Now, instead of having the client talk to
the server directly, communication goes from the client to the client
rproxy to the server rproxy to the server, and vice versa.

browser -------- client rproxy -------- server rproxy -------- server

(I will now use the term _browser_ instead of simply client, so that one
 can rapidly grasp the situation)

When the client first requests something from the server, the GET will
go through the proxy pair to the server and the response will be
forwarded to the browser, and cached in the client rproxy.

![rproxy diagram](https://rproxy.samba.org/flowchart.png "How this looks
    like")

I use the term _client rproxy_ instead of _decoding rproxy_ and _server
rproxy_ instead of _encoding rproxy_, because encoding/decoding is
actually done in the 2 directions so these can be confusing.

When the client re-requests the server for the same resource (1), the client
proxy intercepts the request, reads if there is something similar in the
cache (2). If there is nothing, everything will happen as before. If there
is something, the client rproxy will calculate the rsync signature of
the data and add this (url-base64-encoded-) signature as a
_X-RProxy-Sig_ header in the request to the server (3). The server rproxy
will GET the resource as usual (4), and do some calculation when it receives
the response (5).

Upon receiving the response, the server rproxy has a _new file_ and an
old file _signature_, which allow it to calculate a _delta_ from old
file to new file. It will then send this _delta_ to the client rproxy
(6).

The client rproxy has a _old file_ and a _delta_, so it will patch the
old file to produce _new file_, which it can now serve to the browser
(7) (and also save in the cache for further requests)

In the previous diagram, you can see that the big data exchange was on
fast links, while the slow link (across the internet) just exchanged
little data: goal is achieved.

## When to use

As you can see, this setup needs no modification in the browser nor in
the server, only to set the browser's proxy to the client rproxy and to
set the server rproxy as a frontal to the logic server. However, a
better solution would be to directly integrate these 2 parts in the
respective components for better performance (although there could still
be value in having a domain-wide client rproxy, so that requests to
the same urls from different browsers can be effectively deduplicated).

This functionality allows a server to send (far) less content when said
content is dynamic: I believe this is completely adapted to the current
state of the Web, were most content is dynamic, developers mix endless
AJAX calls to incrementally update a page and everything moves so fast.

On the other hand, rproxy will be a waste of bandwidht and CPU time if
the content is static, for instance for a CDN that distributes static
assets such as images or js scripts. If you have one of these, don't use
something like rproxy.

## Some numbers

The repo currently contains a dummy server that serves two pages:
_index.html_ and _index2.html_, which are 2 very similar versions of the
webpage at www.cnn.com. Whenever a GET is satisfied, one of the 2 pages
is served, and the other one is queued for the next request (so that 2
    consecutive calls basically test the diffing from one to the other).

I have chosen this page because it is very huge (116 kB!), but it is
extremely dynamic: the 2 pages were fetched ~1 min apart, and are a mere
3 bytes (!) different. This means that if you browse to this website,
you will have to download an enormous blob for something you can
merely notice.

With the rsync algorithm (and thus with rproxy), a _signature_ of either
of the page amounts to 708B, which means 944B after base64-encoding, and
the _delta_ produced by the algorithm is 2067B long.

So, instead of downloading 116 kB of data just to get those new 3 bytes,
you will send 944B upstream and receive 2067B of data. That is still
an enormous amount of data, but far less than the original amount, and
you won't be able to do much better unless you start uploading more
(which is a bad idea when you look at the current abysmal asymetry of
 the typical ADSL line)
