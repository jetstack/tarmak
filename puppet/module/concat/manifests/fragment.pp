# == Define: concat::fragment
#
# Creates a concat_fragment in the catalogue
#
# === Options:
#
# [*target*]
#   The file that these fragments belong to
# [*content*]
#   If present puts the content into the file
# [*source*]
#   If content was not specified, use the source
# [*order*]
#   By default all files gets a 10_ prefix in the directory you can set it to
#   anything else using this to influence the order of the content in the file
#
define concat::fragment(
    $target,
    $ensure  = undef,
    $content = undef,
    $source  = undef,
    $order   = '10',
) {
  validate_string($target)
  $resource = 'Concat::Fragment'

  if $ensure != undef {
    warning('The $ensure parameter to concat::fragment is deprecated and has no effect.')
  }

  validate_string($content)
  if !(is_string($source) or is_array($source)) {
    fail("${resource}['${title}']: 'source' is not a String or an Array.")
  }

  if !(is_string($order) or is_integer($order)) {
    fail("${resource}['${title}']: 'order' is not a String or an Integer.")
  } elsif (is_string($order) and $order =~ /[:\n\/]/) {
    fail("${resource}['${title}']: 'order' cannot contain '/', ':', or '\n'.")
  }

  if ! ($content or $source) {
    crit('No content, source or symlink specified')
  } elsif ($content and $source) {
    fail("${resource}['${title}']: Can't use 'source' and 'content' at the same time.")
  }

  $safe_target_name = regsubst($target, '[/:~\n\s\+\*\(\)@]', '_', 'GM')

  concat_fragment { $name:
    target  => $target,
    tag     => $safe_target_name,
    order   => $order,
    content => $content,
    source  => $source,
  }
}
