require 'puppetlabs_spec_helper/rake_tasks'
require 'puppet_blacksmith/rake_tasks'
require 'puppet-lint/tasks/puppet-lint'
require 'metadata-json-lint/rake_task'
require 'rspec/core/rake_task'

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
task :minikube => :librarian_prepare
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

desc 'Run minikube acceptance tests'
task :minikube do
  Rake::Task[:spec_prep].invoke
  Rake::Task[:minikube_standalone].invoke
  Rake::Task[:spec_clean].invoke
end

desc 'Run rspec against minikube only'
RSpec::Core::RakeTask.new(:minikube_standalone) do |t|
  at_exit do
    `minikube delete --profile #{ENV['MINIKUBE_PROFILE']}` unless ENV['MINIKUBE_PROFILE'].nil?
  end
  t.pattern = 'spec/{aliases,classes,defines,unit,functions,hosts,integration,type_aliases,types}/**/*_spec.rb'
  t.rspec_opts = ['--color','--tag','minikube']
  t.rspec_opts << ENV['CI_SPEC_OPTIONS'] unless ENV['CI_SPEC_OPTIONS'].nil?
end

desc 'Redefine spec without minikube tests'
Rake::Task[:spec_standalone].clear
RSpec::Core::RakeTask.new(:spec_standalone) do |t|
  t.pattern = 'spec/{aliases,classes,defines,unit,functions,hosts,integration,type_aliases,types}/**/*_spec.rb'
  t.rspec_opts = ['--color','--tag','~minikube']
  t.rspec_opts << ENV['CI_SPEC_OPTIONS'] unless ENV['CI_SPEC_OPTIONS'].nil?
end
