class tarmak::fluent_bit(
){
  include ::tarmak

  $::tarmak::fluent_bit_configs.each |Integer $index, String $fluent_bit_config| {
    ::fluent_bit::output{"fluent-bit-output-$index":
      config => parsejson($fluent_bit_config),
    }
  }

}
