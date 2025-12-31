## 2024-05-20 - Unauthenticated SOCKS5 Proxy
**Vulnerability:** The SOCKS5 proxy was open to the public, allowing anyone to use it without authentication.
**Learning:** Open proxies can be abused for malicious activities, such as hiding the attacker's identity, bypassing network restrictions, or launching attacks on other systems. The codebase was written with a focus on solving a specific network problem, and security was not a primary consideration.
**Prevention:** Always implement strong authentication for any proxy server to ensure that only authorized users can access it. Use environment variables or a secrets management system to handle credentials securely.
