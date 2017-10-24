# Copyright Jetstack Ltd. See LICENSE for details.

verify: terraform_fmt terraform_validate

terraform_fmt:
	git ls-files **/*.tf | xargs -n1 dirname | sort -u | xargs -n 1 terraform fmt -write=false -diff=true -check=true

terraform_validate:
	git ls-files **/*.tf | xargs -n1 dirname | sort -u | xargs -n 1 terraform validate -check-variables=false

