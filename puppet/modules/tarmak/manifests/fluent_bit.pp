class tarmak::fluent_bit(
){
  include ::tarmak

  $::tarmak::fluent_bit_configs.each |Integer $index, Hash $fluent_bit_config| {
    ::fluent_bit::output{"fluent-bit-output-$index":
      config => $fluent_bit_config,
    }
  }

}
