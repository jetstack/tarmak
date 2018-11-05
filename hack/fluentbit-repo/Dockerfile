FROM centos:7

RUN yum install -y epel-release && yum -y install tinyproxy

ENV http_proxy "http://127.0.0.1:8888"

ENV repo_url "http://packages.fluentbit.io/centos/7"

RUN \
  tinyproxy && \
  echo -e "[td-agent-bit]\nname = TD Agent Bit\nbaseurl = ${repo_url}\ngpgcheck=1\ngpgkey=http://packages.fluentbit.io/fluentbit.key\nenabled=1\n" > /etc/yum.repos.d/td-agent-bit.repo && \
  yum -y install td-agent-bit

ENV http_proxy ""

RUN set -x; set -o pipefail; set -e; for url in $(cat /var/log/tinyproxy/tinyproxy.log  | grep -oP "${repo_url}/([^ ]*)"); do dest="/output${url//${repo_url}/}" ; mkdir -p "$(dirname "${dest}")"; curl -sL -o "${dest}" "${url}"; done
