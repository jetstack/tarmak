Path too long for Unix domain socket
------------------------------------

If you run into the following error when running ``tarmak clusters apply`` or another Tarmak command:

::

  unix_listener: path "/var/folders/1h/11bqph8n0wl43my28ptqdp7m0000gn/T//ssh-control-centos@x.x.x.x:22.RiuGqfWBGp2rhXlp" too long for Unix domain socket

In certain cases (e.g. MacOS) it is possible that the temp directory path is too long. This can be solved by changing the temp directory path to a shorter one:

::

  export TMPDIR=/tmp
