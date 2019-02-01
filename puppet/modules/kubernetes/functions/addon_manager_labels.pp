function kubernetes::addon_manager_labels(String $manifests) {
  $reconcile_label = 'addonmanager.kubernetes.io/mode: Reconcile'
  $ensure_exists_label = 'addonmanager.kubernetes.io/mode: EnsureExists'

  $yaml_array = split($manifests, '---')

  $yaml_array.each |String $yaml| {
    if $yaml.strip != '' {
      $yaml_obj = parseyaml($yaml)
      $addon_manager_label_value = $yaml_obj.dig('metadata','labels','addonmanager.kubernetes.io/mode')
      if $addon_manager_label_value == undef {
        fail("yaml did not contain addon manager label: ${yaml}")
      }
      if $addon_manager_label_value != 'Reconcile' and $addon_manager_label_value != 'EnsureExists' {
        fail("yaml did contain incorrect addon manager label value: ${addon_manager_label_value}")
      }
    }
  }
}
