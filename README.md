# `gpdns`

A simple drop-in replacement for `miekg/dns.Client.Exchange` using the Google
Public DNS HTTPS API.

While `gpdns.Client.Exchange` accepts a `dns.Msg` most fancy features of the normal
DNS client are not actually supported. Refer to the Google API documentation for
information on the kind of queries that can be made.

API specification: https://developers.google.com/speed/public-dns/docs/dns-over-https#api_specification
