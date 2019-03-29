#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

JOB_PATH=.
JOB_TARGET=""
PUPPET_JOB=0

parse_job_name () {
    if [[ $1 == tarmak-puppet-* ]]; then
        PUPPET_JOB=1
    fi
    if [[ $1 == tarmak-puppet-module-* ]]; then
        postfix=${1#tarmak-puppet-module-}
        parts=(${postfix//-/ })
        JOB_PATH=puppet/modules/${parts[0]}
        JOB_TARGET=${postfix#${parts[0]}-}
    fi
    if [[ $1 == tarmak-puppet-roles-* ]]; then
        JOB_TARGET=${1#tarmak-puppet-roles-}
        JOB_PATH=puppet
    fi
    if [[ $JOB_TARGET == "quick-verify" ]]; then
        JOB_TARGET=verify
    fi
}

assert () {
    if [[ "$1" != "$2" ]]; then
        echo "unexpected value actual=$1 expected=$2"
        exit 1
    fi
}

test_job_names () {
    parse_job_name tarmak-puppet-roles-quick-verify
    assert $JOB_PATH puppet
    assert $JOB_TARGET verify
    assert $PUPPET_JOB 1

    parse_job_name tarmak-puppet-module-aws_ebs-quick-verify
    assert $JOB_PATH puppet/modules/aws_ebs
    assert $JOB_TARGET verify
    assert $PUPPET_JOB 1

    parse_job_name tarmak-puppet-module-etcd-acceptance-single-node
    assert $JOB_PATH puppet/modules/etcd
    assert $JOB_TARGET acceptance-single-node
    assert $PUPPET_JOB 1

    parse_job_name tarmak-puppet-module-etcd-acceptance-three-node
    assert $JOB_PATH puppet/modules/etcd
    assert $JOB_TARGET acceptance-three-node
    assert $PUPPET_JOB 1

    parse_job_name tarmak-puppet-module-etcd-quick-verify
    assert $JOB_PATH puppet/modules/etcd
    assert $JOB_TARGET verify
    assert $PUPPET_JOB 1
}

## run tests if requested
if [[ -v 1 ]] && [[ $1 == "test" ]]; then
    test_job_names
    exit 0
fi

parse_job_name ${JOB_NAME}

# test if local fixtures files exist
if [[ "$PUPPET_JOB" == "1" ]]; then
    if test -e "${JOB_PATH}/.fixtures.yml.local"; then
        export FIXTURES_YML=.fixtures.yml.local
    fi
fi

make -C ${JOB_PATH} ${JOB_TARGET}
