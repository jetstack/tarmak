function kubernetes::addon_manager_labels(String $manifests) {
  $reconcile_label = 'addonmanager.kubernetes.io/mode: Reconcile'
  $ensure_exists_label = 'addonmanager.kubernetes.io/mode: EnsureExists'
  $yaml_array = split($manifests, '---')

  $yaml_array.each |String $yaml| {
    if $yaml.strip != '' and split($yaml, '\n').length > 1 {
      if !($yaml =~ $reconcile_label or $yaml =~ $ensure_exists_label) {
        fail("yaml did not contain addon manager label: ${yaml}")
      }
    }
  }
}
