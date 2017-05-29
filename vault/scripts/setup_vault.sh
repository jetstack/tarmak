#!/bin/bash
# script to set up Vault for Kubernetes
set -e

echo "################################"
echo "#### --- Setting up PKI --- ####"
echo "################################"
echo "Pointing at ${VAULT_ADDR}"

COMPONENTS="etcd-k8s etcd-overlay k8s"
BASE_PATH="${CLUSTER_ID}/pki"


# create a CA per component (not intermediate)
for component in ${COMPONENTS}; do
  path="${BASE_PATH}/${component}"
  description="Kubernetes ${CLUSTER_ID}/${component} CA"

  # test for existence of mount in list of mounts
  if [[ "$(vault mounts | grep ^${path}\/ | wc -l)" -eq 0 ]]; then
    # mount component
    vault mount -path "${path}" -description "${description}" pki
    vault mount-tune -max-lease-ttl=87600h "${path}"
  fi
  # test for existence of ca at path
  if [[ "$(vault read ${path}/cert/ca | grep 'BEGIN CERTIFICATE' | wc -l)" -eq 0 ]]; then
    # generate a new CA
    vault write "${path}/root/generate/internal" \
      common_name="${description}" \
      ttl=87600h # 10 years
  fi
  # if it's a etcd ca populate only a single role
  if [[ "${component}" == etcd-* ]]; then
    vault write "${path}/roles/client" \
      use_csr_common_name=false \
      allow_any_name=true \
      max_ttl="720h" \
      allow_ip_sans=true \
      server_flag=true \
      client_flag=true
    vault write "${path}/roles/server" \
      use_csr_common_name=false \
      use_csr_sans=false \
      allow_any_name=true \
      max_ttl="720h" \
      allow_ip_sans=true \
      server_flag=true \
      client_flag=true
  fi
  # if it's k8s
  if [[ "${component}" == "k8s" ]]; then
    vault write "${path}/roles/admin" \
      use_csr_common_name=false \
      allowed_domains="admin" \
      allow_bare_domains=true \
      allow_localhost=false \
      allow_subdomains=false \
      allow_ip_sans=false \
      server_flag=false \
      client_flag=true \
      max_ttl="720h"
    for role in kube-scheduler kube-controller-manager kube-proxy; do
      vault write "${path}/roles/${role}" \
        use_csr_common_name=false \
        allowed_domains="${role},system:${role}" \
        allow_bare_domains=true \
        allow_localhost=false \
        allow_subdomains=false \
        allow_ip_sans=false \
        server_flag=false \
        client_flag=true \
        max_ttl="720h"
    done
      vault write "${path}/roles/kubelet" \
        use_csr_common_name=false \
        use_csr_sans=false \
        allowed_domains="kubelet,system:node:*" \
        allow_bare_domains=true \
        allow_glob_domains=true \
        allow_localhost=false \
        allow_subdomains=false \
        server_flag=true \
        client_flag=true \
        max_ttl="720h"
      vault write "${path}/roles/kube-apiserver" \
        use_csr_common_name=false \
        use_csr_sans=false \
        allow_localhost=true \
        allow_any_name=true \
        allow_bare_domains=true \
        allow_ip_sans=true \
        server_flag=true \
        client_flag=false \
        max_ttl="720h"
  fi
done

# Generic secrets mount
secrets_path="${CLUSTER_ID}/secrets"
# test for existence of secrets mount in list of mounts
if [[ "$(vault mounts | grep ^${CLUSTER_ID}/secrets\/ | wc -l)" -eq 0 ]]; then
  vault mount -path "${secrets_path}" -description="Kubernetes ${CLUSTER_ID} secrets" generic
  # Generate a key for the service accounts
  openssl genrsa 4096 | vault write "${secrets_path}/service-accounts" key=-
fi

rm -f ${CLUSTER_ID}-token
rm -f tokens.tfvar

# Generate policies per node role
for role in master worker etcd admin; do
  policy_name="${CLUSTER_ID}/${role}"
  policy=""
  if [[ "${role}" == "master" ]] || [[ "${role}" == "worker" ]]; then
    for cert_role in k8s/sign/kubelet k8s/sign/kube-proxy etcd-overlay/sign/client; do
      policy="${policy}
path \"${BASE_PATH}/${cert_role}\" {
    capabilities = [\"create\",\"read\",\"update\"]
}
"
     done
   fi
   if [[ "${role}" == "master" ]]; then
     for cert_role in k8s/sign/kube-apiserver k8s/sign/kube-scheduler k8s/sign/kube-controller-manager k8s/sign/admin etcd-k8s/sign/client; do
       policy="${policy}
path \"${BASE_PATH}/${cert_role}\" {
    capabilities = [\"create\",\"read\",\"update\"]
}
"
     done

   policy="${policy}
path \"${secrets_path}/service-accounts\" {
    capabilities = [\"read\"]
}
"
    fi
    if [[ "${role}" == "etcd" ]]; then
      for cert_role in etcd-k8s/sign/server etcd-overlay/sign/server; do
        policy="${policy}
path \"${BASE_PATH}/${cert_role}\" {
    capabilities = [\"create\",\"read\",\"update\"]
}
"
      done
    fi

    if [[ "${role}" == "admin" ]]; then
      for cert_role in k8s/sign/admin; do
        policy="${policy}
path \"${BASE_PATH}/${cert_role}\" {
    capabilities = [\"create\",\"read\",\"update\"]
}
"
      done
    fi

  # write out new policy idempotently
  echo "${policy}" | vault policy-write "${policy_name}" -

  # test for existence of token role at specified path and grep for policy name (twice)
  if [[ "$(vault read auth/token/roles/${CLUSTER_ID}-${role} | grep $policy_name | wc -l)" -lt 2 ]]; then
    # create token role
    token_role="auth/token/roles/${CLUSTER_ID}-${role}"
    vault write "${token_role}" \
      period="720h" \
      orphan=true \
      allowed_policies="default,${policy_name}" \
      path_suffix="${policy_name}"
  fi

  # test for existence of token creator policy in (same) list of policies
  if [[ "$(vault policies | grep ^${policy_name}-creator\$ | wc -l)" -eq 0 ]]; then
    # create token creator policy
    vault policy-write "${policy_name}-creator" - <<EOF
path "auth/token/create/${CLUSTER_ID}-${role}" {
    capabilities = ["create","read","update"]
}
EOF
  fi

  # test for existence of tokens (indicated by an entry in the secrets path)
  if vault read -field init_token "${secrets_path}/init-token-${role}" 2>/dev/null > /dev/null; then
    init_token="$(vault read -field init_token "${secrets_path}/init-token-${role}")"
  else
    #echo "Token init-token-${role} is missing, generating a new one!"
    init_token=$(vault token-create \
      -display-name="${policy_name}-creator" \
      -format json \
      -orphan \
      -ttl="8760h" \
      -period="8760h" \
      -policy="${policy_name}-creator" | jq -r .auth.client_token
    )
    vault write "${secrets_path}/init-token-${role}" "init_token=${init_token}" > /dev/null
  fi
  # print out init tokens - format TBC
  if [[ "${role}" == "admin" ]]; then
    echo $init_token >> ${CLUSTER_ID}-token
  else
    echo "vault_init_token_${role}=\"${init_token}\"" >> tokens.tfvar
  fi
done
