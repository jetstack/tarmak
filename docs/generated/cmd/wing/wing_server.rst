.. _wing_server:

wing server
-----------

Launch a wing API server

Synopsis
~~~~~~~~


Launch a wing API server

::

  wing server [flags]

Options
~~~~~~~

::

      --bind-address ip                          The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank, all interfaces will be used (0.0.0.0 for all IPv4 interfaces and :: for all IPv6 interfaces). (default 0.0.0.0)
      --cert-dir string                          The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored. (default "apiserver.local.config/certificates")
      --default-watch-cache-size int             Default watch cache size. If zero, watch cache will be disabled for resources that do not have a default watch size set. (default 100)
      --delete-collection-workers int            Number of workers spawned for DeleteCollection call. These are used to speed up namespace cleanup. (default 1)
      --enable-garbage-collector                 Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-controller-manager. (default true)
      --encryption-provider-config string        The file containing configuration for encryption providers to be used for storing secrets in etcd
      --environment string                       this specifies the environment name
      --etcd-cafile string                       SSL Certificate Authority file used to secure etcd communication.
      --etcd-certfile string                     SSL certification file used to secure etcd communication.
      --etcd-compaction-interval duration        The interval of compaction requests. If 0, the compaction request from apiserver is disabled. (default 5m0s)
      --etcd-count-metric-poll-period duration   Frequency of polling etcd for number of resources per type. 0 disables the metric collection. (default 1m0s)
      --etcd-keyfile string                      SSL key file used to secure etcd communication.
      --etcd-prefix string                       The prefix to prepend to all resource paths in etcd. (default "/registry/wing.tarmak.io")
      --etcd-servers strings                     List of etcd servers to connect with (scheme://ip:port), comma separated.
      --etcd-servers-overrides strings           Per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are URLs, semicolon separated.
  -h, --help                                     help for server
      --http2-max-streams-per-connection int     The limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default. (default 1000)
      --secure-port int                          The port on which to serve HTTPS with authentication and authorization.If 0, don't serve HTTPS at all. (default 443)
      --storage-backend string                   The storage backend for persistence. Options: 'etcd3' (default).
      --storage-media-type string                The media type to use to store objects in storage. Some resources or storage backends may only support a specific media type and will ignore this setting. (default "application/json")
      --tls-cert-file string                     File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
      --tls-cipher-suites strings                Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be use.  Possible values: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_RC4_128_SHA,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA256,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_RC4_128_SHA
      --tls-min-version string                   Minimum TLS version supported. Possible values: VersionTLS10, VersionTLS11, VersionTLS12
      --tls-private-key-file string              File containing the default x509 private key matching --tls-cert-file.
      --tls-sni-cert-key namedCertKey            A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed wildcard segments. If no domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com". (default [])
      --watch-cache                              Enable watch caching in the apiserver (default true)
      --watch-cache-sizes strings                List of watch cache sizes for every resource (pods, nodes, etc.), comma separated. The individual override format: resource[.group]#size, where resource is lowercase plural (no version), group is optional, and size is a number. It takes effect when watch-cache is enabled. Some resources (replicationcontrollers, endpoints, nodes, pods, services, apiservices.apiregistration.k8s.io) have system defaults set by heuristics, others default to default-watch-cache-size

SEE ALSO
~~~~~~~~

* `wing <wing.html>`_ 	 - wing is the agent component for tarmak, it runs on every instance of tarmak

