class calico::worker
{
  class {'calico::bin_install':} ->
  class {'calico::lo_install':} ->
  class {'calico::config':} ->
  class {'calico::node':}
}
