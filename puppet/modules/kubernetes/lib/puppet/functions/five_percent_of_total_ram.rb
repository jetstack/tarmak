Puppet::Functions.create_function(:five_percent_of_total_ram) do
  dispatch :five_percent_of_total_ram do
    param 'Integer', :total_bytes
  end

  def five_percent_of_total_ram(total_bytes)
    five_percent = (total_bytes * 0.05 / 1024**2).round
    default = 100

    if five_percent < default
      '100Mi'
    else
      five_percent.to_s + 'Mi'
    end
  end
end
