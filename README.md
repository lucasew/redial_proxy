# redial_proxy

Sometimes the internet here is a bit buggy, that kind of bug like sites not loading CSS at the first try or even the
browser showing a error like when you are without internet.

This is my workaround try to solve my case and maybe yours too.

This is simple enough to fit in only one file.

Its basically a SOCKS5 proxy server with a custom dialer, that is the guy who is responsible to establishing the connection, 
my custom dialer is a wrapper around the default dialer that retries the connection after 100ms if the error have route in its message.

This is my best bet without working with the anoying bureaucracy with people that are not solving my problem.

## Installing

- Mise

```
mise use github:lucasew/redial_proxy
```
