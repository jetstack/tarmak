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

      --admission-control strings                        Admission is divided into two phases. In the first phase, only mutating admission plugins run. In the second phase, only validating admission plugins run. The names in the below list may represent a validating plugin, a mutating plugin, or both. Within each phase, the plugins will run in the order in which they are passed to this flag. Comma-delimited list of: Initializers, InstanceInitTime, MutatingAdmissionWebhook, NamespaceLifecycle, ValidatingAdmissionWebhook. (default [InstanceInitTime])
      --admission-control-config-file string             File with admission control configuration.
      --bind-address ip                                  The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank, all interfaces will be used (0.0.0.0). (default 0.0.0.0)
      --cert-dir string                                  The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored. (default "apiserver.local.config/certificates")
      --default-watch-cache-size int                     Default watch cache size. If zero, watch cache will be disabled for resources that do not have a default watch size set. (default 100)
      --delete-collection-workers int                    Number of workers spawned for DeleteCollection call. These are used to speed up namespace cleanup. (default 1)
      --deserialization-cache-size int                   Number of deserialized json objects to cache in memory.
      --enable-garbage-collector                         Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-controller-manager. (default true)
      --etcd-cafile string                               SSL Certificate Authority file used to secure etcd communication.
      --etcd-certfile string                             SSL certification file used to secure etcd communication.
      --etcd-compaction-interval duration                The interval of compaction requests. If 0, the compaction request from apiserver is disabled. (default 5m0s)
      --etcd-keyfile string                              SSL key file used to secure etcd communication.
      --etcd-prefix string                               The prefix to prepend to all resource paths in etcd. (default "/registry/wing.tarmak.io")
      --etcd-servers strings                             List of etcd servers to connect with (scheme://ip:port), comma separated.
      --etcd-servers-overrides strings                   Per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are http://ip:port, semicolon separated.
      --experimental-encryption-provider-config string   The file containing configuration for encryption providers to be used for storing secrets in etcd
  -h, --help                                             help for server
      --secure-port int                                  The port on which to serve HTTPS with authentication and authorization. If 0, don't serve HTTPS at all. (default 443)
      --storage-backend string                           The storage backend for persistence. Options: 'etcd3' (default), 'etcd2'.
      --storage-media-type string                        The media type to use to store objects in storage. Some resources or storage backends may only support a specific media type and will ignore this setting. (default "application/json")
      --tls-ca-file string                               If set, this certificate authority will used for secure access from Admission Controllers. This must be a valid PEM-encoded CA bundle. Altneratively, the certificate authority can be appended to the certificate provided by --tls-cert-file.
      --tls-cert-file string                             File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
      --tls-private-key-file string                      File containing the default x509 private key matching --tls-cert-file.
      --tls-sni-cert-key namedCertKey                    A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed wildcard segments. If no domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com". (default [])
      --watch-cache                                      Enable watch caching in the apiserver (default true)
      --watch-cache-sizes strings                        List of watch cache sizes for every resource (pods, nodes, etc.), comma separated. The individual override format: resource#size, where size is a number. It takes effect when watch-cache is enabled.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)

SEE ALSO
~~~~~~~~

* `wing <wing.html>`_ 	 - wing is the agent component for tarmak, it runs on every instance of tarmak

