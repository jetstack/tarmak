Puppet::Type.type(:archive).provide(:airworthy, parent: :ruby) do
  commands airworthy: 'airworthy'

  def airworthy_params(params)
    params += ['--signature-armored', resource[:signature_armored]] if resource[:signature_armored]
    params += ['--signature-binary', resource[:signature_binary]] if resource[:signature_binary]
    params += ['--sha256sums', resource[:sha256sums]] if resource[:sha256sums]
    params
  end

  def download(filepath)
    params = airworthy_params(
      [
        'download',
        '--output', filepath,
      ]
    )
    params << resource[:source]

    airworthy(params)
  end
end
