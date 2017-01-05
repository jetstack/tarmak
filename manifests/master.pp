class calico::master
{
  class {'calico::bin_install':} ->
  class {'calico::lo_install':} ->
  class {'calico::config':} ->
  class {'calico::node':} ->
  class {'calico::policy_controller':}
}
