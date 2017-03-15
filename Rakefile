require 'puppetlabs_spec_helper/rake_tasks'
require 'puppet_blacksmith/rake_tasks'
require 'puppet-lint/tasks/puppet-lint'
require 'metadata-json-lint/rake_task'

if RUBY_VERSION >= '1.9'
  require 'rubocop/rake_task'
  RuboCop::RakeTask.new
end

PuppetLint.configuration.send('disable_80chars')
PuppetLint.configuration.relative = true
PuppetLint.configuration.ignore_paths = ['spec/**/*.pp', 'pkg/**/*.pp']

desc 'Prepare module dependecies'
task :librarian_prepare do
  sh 'librarian-puppet install --path=spec/fixtures/modules'
end

desc 'Clean module dependecies'
task :librarian_clean do
  sh 'librarian-puppet clean'
end
task :spec => :librarian_prepare
task :clean => :librarian_clean

task :beaker => :librarian_prepare

desc 'Validate manifests, templates, and ruby files'
task :validate do
  Dir['manifests/**/*.pp'].each do |manifest|
    sh "puppet parser validate --noop #{manifest}"
  end
  Dir['spec/**/*.rb', 'lib/**/*.rb'].each do |ruby_file|
    sh "ruby -c #{ruby_file}" unless ruby_file =~ %r{spec/fixtures}
  end
  Dir['templates/**/*.erb'].each do |template|
    sh "erb -P -x -T '-' #{template} | ruby -c"
  end
end

desc 'Run metadata_lint, lint, validate, and spec tests.'
task :test do
  [:metadata_lint, :lint, :validate, :spec].each do |test|
    Rake::Task[test].invoke
  end
end

desc 'Generate README.md'
task :readme do
   sh 'puppet strings generate --emit-json documentation.json'
   metadata = JSON.parse(File.open('metadata.json').read)
   doc = JSON.parse(File.open('documentation.json').read)
   puppet_module = metadata['name'].split('-')[1]
   main_class = nil
   doc['puppet_classes'].each do |c|
     main_class = c if c['name'] == puppet_module
   end

   output = []
   output << "# #{puppet_module}"
   output << ""
   output << "#### Table of Contents"
   output << ""
   output << "1. [Description](#description)"
   output << "2. [Classes](#classes)"
   output << "3. [Defined Types](#defined-types)"
   output << ""
   output << "## Description"
   output << ""
   output << main_class['docstring']['text']

   output << "## Classes"
   output << ""
   doc['puppet_classes'].each do |c|
      output << "### `#{c['name']}`\n"
      output << c['docstring']['text']
      output << ""

      output << "#### Parameters\n"
      c['docstring']['tags'].each do |t|
        next unless t['tag_name'] == 'param'
        output << "##### `#{t['name']}`\n"
        output << "* #{t['text']}"
        output << "* Type: `#{t['types']}`"
        output << "* Default: `#{c['defaults'][t['name']]}`"
        output << ''
      end

      output << "#### Examples\n"
      c['docstring']['tags'].each do |t|
        next unless t['tag_name'] == 'example'
        output << "##### #{t['name']}\n"
        output << "```\n#{t['text']}\n```"
      end
   end
  open('README.md', 'w') do |f|
    f.puts output.join("\n")
  end
end
