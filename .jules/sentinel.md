## 2026-01-20 - Add Basic Authentication to SOCKS5 Proxy

**Vulnerability:** The SOCKS5 proxy server currently lacks any authentication mechanism. Even when bound to the loopback interface, this allows any user on the local machine (or anyone with network access if the binding is changed) to use the proxy without restriction.
**Learning:** Security defaults are crucial. While the tool is intended for local use, "defense in depth" suggests adding an authentication layer to prevent unauthorized usage, especially in multi-user environments.
**Prevention:** Implement support for standard SOCKS5 username/password authentication controlled via environment variables.
